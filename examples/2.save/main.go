package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	contractevent "github.com/lengzhao/contract_event"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	RCPNode       string                           `yaml:"rpc_node"`
	Subscriptions []contractevent.SubscriptionConf `yaml:"subscriptions"`
}

func main() {
	confFile := flag.String("conf", "../config.yaml", "config file(yaml)")
	dbFile := flag.String("db", "sqlite.db", "sqlite file")
	flag.Parse()
	data, err := os.ReadFile(*confFile)
	if err != nil {
		log.Fatal("fail to read config file:", *confFile, err)
	}
	var conf Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal("fail to unmarshal config:", err)
	}
	db, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		log.Fatal("fail to open sqlite file:", err)
	}
	defer func() {
		ldb, err := db.DB()
		if err == nil {
			ldb.Close()
		}
	}()
	client, err := ethclient.Dial(conf.RCPNode)
	if err != nil {
		log.Println("fail to do ethclient.Dial:", err)
		os.Exit(2)
	}
	for _, sub := range conf.Subscriptions {
		contractevent.CreateTable(db, sub.Alias)
		event, err := contractevent.NewEvent(sub, client, func(alias string, info map[string]interface{}) error {
			var item contractevent.DBItem
			item.TX = info[contractevent.KTX].(string)
			item.LogIndex = info[contractevent.KLogIndex].(uint)
			item.Others, _ = json.Marshal(info)
			id, err := contractevent.InsertItem(db, alias, item)
			log.Println("new event:", alias, id, item.TX, item.LogIndex)
			return err
		})
		if err != nil {
			log.Println("fail to sub event:", err)
			os.Exit(2)
		}
		err = event.Run(sub.StartBlock, sub.StartBlock+1)
		if err != nil {
			log.Println("fail to run:", err)
			os.Exit(2)
		}
	}
}
