package contractevent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Event struct {
	conf   SubscriptionConf
	cb     EventCallback
	query  ethereum.FilterQuery
	eABI   abi.ABI
	client *ethclient.Client
}

type EventCallback func(alias string, info map[string]interface{}) error

const (
	KAlias       = "alias"
	KBlock       = "block"
	KBlockNumber = "block_number"
	KContract    = "contract"
	KTX          = "tx"
	KLogIndex    = "log_index"
	KTopic       = "topic"
	KEventName   = "event_name"
	KRawData     = "raw_data"
)

func NewEventWithDB(conf SubscriptionConf, client *ethclient.Client, db *gorm.DB) (*Event, error) {
	err := CreateEventTable(db, conf.Alias)
	if err != nil {
		log.Warnln("fail to create database table of event ", conf.Alias, err)
	}
	return NewEvent(conf, client, func(alias string, info map[string]interface{}) error {
		var item DBItem
		item.TX = info[KTX].(string)
		item.LogIndex = info[KLogIndex].(uint)
		item.Others, _ = json.Marshal(info)
		id, err := InsertItem(db, alias, item)
		log.Infoln("new event:", alias, id, item.TX, item.LogIndex, err)
		return err
	})
}

func NewEvent(conf SubscriptionConf, client *ethclient.Client, cb EventCallback) (*Event, error) {
	var out Event
	out.conf = conf
	out.cb = cb
	data, err := os.ReadFile(conf.ABIFile)
	if err != nil {
		data = GetABIData(conf.ABIFile)
		if len(data) == 0 {
			log.Errorln("fail to open abi file:", conf.ABIFile, err)
			return nil, err
		}
	}
	cAbi, err := abi.JSON(bytes.NewReader(data))
	if err != nil {
		log.Errorln("fail to load abi:", conf.ABIFile, err)
		return nil, err
	}

	out.query, err = newQuery(conf, cAbi)
	if err != nil {
		return nil, err
	}

	for _, event := range cAbi.Events {
		if _, ok := cAbi.Events[event.ID.Hex()]; ok {
			continue
		}
		for i, it := range event.Inputs {
			if it.Indexed {
				it.Indexed = false
				event.Inputs[i] = it
			}
		}
		cAbi.Events[event.ID.Hex()] = event
	}

	out.eABI = cAbi
	out.client = client

	return &out, nil
}

func (e *Event) Run(start, end uint64) error {
	query := e.query
	query.FromBlock = new(big.Int).SetUint64(start)
	query.ToBlock = new(big.Int).SetUint64(end)
	logs, err := e.client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Errorln("fail to FilterLogs:", e.conf.Alias, err)
		return err
	}

	for _, vLog := range logs {
		tid := vLog.Topics[0].Hex()
		info := make(map[string]interface{})
		info[KAlias] = e.conf.Alias
		info[KContract] = vLog.Address.Hex()
		info[KBlock] = vLog.BlockHash.Hex()
		info[KBlockNumber] = vLog.BlockNumber
		info[KTX] = vLog.TxHash.Hex()
		info[KLogIndex] = vLog.Index
		info[KTopic] = tid
		info[KEventName] = e.eABI.Events[tid].Name
		var data []byte
		for i, t := range vLog.Topics {
			if i == 0 {
				continue
			}
			data = append(data, t.Bytes()...)
		}
		data = append(data, vLog.Data...)
		err := e.eABI.UnpackIntoMap(info, tid, data)
		if err != nil {
			log.Warnln("fail to UnpackIntoMap:", e.conf.Alias, info[KEventName], tid, len(data), err)
			info[KRawData] = data
		}
		if len(e.conf.Filter) == 0 {
			err = e.cb(e.conf.Alias, info)
			if err != nil {
				return err
			}
			continue
		}
		err = check(e.conf.Filter, info)
		if err != nil {
			log.Infoln("filter limit:", err)
			continue
		}
		err = e.cb(e.conf.Alias, info)
		if err != nil {
			return err
		}
	}
	return nil
}

func check(filter map[string]string, info map[string]interface{}) error {
	if len(filter) == 0 {
		return nil
	}
	for key, value := range filter {
		v, ok := info[key]
		if !ok {
			log.Debugln("filter check fail, hope exist the key:", key)
			return fmt.Errorf("not exist key:%s", key)
		}
		bVal, _ := json.Marshal(v)

		if bVal[0] == '"' {
			bVal = bVal[1 : len(bVal)-1]
		}
		if value != string(bVal) {
			log.Debugf("filter check fail, different value, hope:%s,get:%s", value, bVal)
			return fmt.Errorf("hope:%s,get:%s", value, bVal)
		}
	}
	return nil
}

func newQuery(conf SubscriptionConf, cAbi abi.ABI) (ethereum.FilterQuery, error) {
	query := ethereum.FilterQuery{}
	for _, addr := range conf.Contract {
		query.Addresses = append(query.Addresses, common.HexToAddress(addr))
	}

	e, ok := cAbi.Events[conf.EventName]
	if !ok {
		// 如果EventName为空，则表示监听合约的所有事件
		if conf.EventName == "" {
			return query, nil
		}
		// 如果非空，且没有找到，说明ABI文件有问题，没有对应事件的ABI
		log.Errorln("not found the Event Name from ABI:", conf.Alias, conf.EventName)
		return query, fmt.Errorf("not found the Event Name from ABI:%s", conf.EventName)
	}
	// 只监听合约的指定事件
	query.Topics = append(query.Topics, []common.Hash{e.ID})
	// 如果有filter，它对应的链上事件有indexed修饰，则可以直接添加到topics里，更确定性的过滤事件
	if len(conf.Filter) > 0 {
		for _, it := range e.Inputs {
			if !it.Indexed {
				break
			}
			fv := conf.Filter[it.Name]
			if fv == "" {
				query.Topics = append(query.Topics, []common.Hash{})
				continue
			}
			if strings.HasPrefix(fv, "0x") {
				query.Topics = append(query.Topics, []common.Hash{common.HexToHash(fv)})
				continue
			}
			switch it.Type.T {
			case abi.IntTy, abi.UintTy:
				// 可能的场景：监听NFT的指定的tokenId的事件
				bv, ok := new(big.Int).SetString(fv, 10)
				if !ok {
					log.Errorln("error filter value,unable parse to big.int:", conf.Alias, it.Name, fv)
					return query, fmt.Errorf("error filter value,key:%s,hope int value:%s", it.Name, fv)
				}
				query.Topics = append(query.Topics, []common.Hash{common.BigToHash(bv)})
			default:
				query.Topics = append(query.Topics, []common.Hash{common.HexToHash(fv)})
			}
		}
	}
	return query, nil
}
