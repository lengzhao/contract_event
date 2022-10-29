package contractevent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Event struct {
	conf   SubscriptionConf
	cb     EventCallback
	query  ethereum.FilterQuery
	eABI   abi.ABI
	ent    abi.Event
	client *ethclient.Client
}

type EventCallback func(alias string, info map[string]interface{}) error

const (
	KBlock       = "block"
	KBlockNumber = "block_number"
	KTX          = "tx"
	KLogIndex    = "log_index"
	KTopic       = "topic"
	KEventName   = "event_name"
	KData        = "data"
)

func NewEvent(conf SubscriptionConf, client *ethclient.Client, cb EventCallback) (*Event, error) {
	var out Event
	out.conf = conf
	out.cb = cb
	out.query = ethereum.FilterQuery{
		Addresses: []common.Address{
			common.HexToAddress(conf.Contract),
		},
	}
	data, err := os.ReadFile(conf.ABIFile)
	if err != nil {
		log.Println("fail to open abi file:", conf.ABIFile, err)
		return nil, err
	}
	cAbi, err := abi.JSON(bytes.NewReader(data))
	if err != nil {
		log.Println("fail to load abi:", conf.ABIFile, err)
		return nil, err
	}
	e, ok := cAbi.Events[conf.EventName]
	if ok {
		out.query.Topics = append(out.query.Topics, []common.Hash{e.ID})
		for i, it := range e.Inputs {
			if it.Indexed {
				it.Indexed = false
				e.Inputs[i] = it
			}
		}
		cAbi.Events[e.ID.Hex()] = e
	} else if conf.EventName != "" {
		log.Println("warning, not found the event name:", conf.Alias, conf.ABIFile, conf.EventName)
		return nil, fmt.Errorf("not found the Event Name from ABI:%s", conf.EventName)
	}

	out.eABI = cAbi
	out.ent = e
	out.client = client

	return &out, nil
}

func (e *Event) Run(start, end uint64) error {
	query := e.query
	query.FromBlock = new(big.Int).SetUint64(start)
	query.ToBlock = new(big.Int).SetUint64(end)
	logs, err := e.client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Println("fail to FilterLogs:", e.conf.Alias, err)
		return err
	}

	for _, vLog := range logs {
		tid := vLog.Topics[0].Hex()
		info := make(map[string]interface{})
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
		err := e.eABI.UnpackIntoMap(info, tid, data)
		if err != nil {
			log.Println("warning, fail to UnpackIntoMap:", e.conf.Alias, tid, err)
			info[KData] = data
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
			log.Println("filter limit:", err)
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
			return fmt.Errorf("not exist key:%s", key)
		}
		bVal, _ := json.Marshal(v)

		if bVal[0] == '"' {
			bVal = bVal[1 : len(bVal)-1]
		}
		if value != string(bVal) {
			return fmt.Errorf("hope:%s,get:%s", value, bVal)
		}
	}
	return nil
}
