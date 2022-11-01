package main

import (
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	contractevent "github.com/lengzhao/contract_event"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RCPNode       string                           `yaml:"rpc_node"`
	Subscriptions []contractevent.SubscriptionConf `yaml:"subscriptions"`
	DB            contractevent.DBConf             `yaml:"db"`
}

func main() {
	confFile := flag.String("conf", "./config.yaml", "config file(yaml)")
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
	db, err := contractevent.NewDB(conf.DB)
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
		contractevent.CreateEventTable(db, sub.Alias)
		event, err := contractevent.NewEventWithDB(sub, client, db)
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
