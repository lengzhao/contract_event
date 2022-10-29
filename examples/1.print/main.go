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
}

func main() {
	confFile := flag.String("conf", "../config.yaml", "config file(yaml)")
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
	client, err := ethclient.Dial(conf.RCPNode)
	if err != nil {
		log.Fatal("fail to do ethclient.Dial:", err)
	}
	for _, sub := range conf.Subscriptions {
		event, err := contractevent.NewEvent(sub, client, func(alias string, info map[string]interface{}) error {
			log.Println("new event:", alias, info)
			return nil
		})
		if err != nil {
			log.Fatal("fail to sub event:", err)
		}
		err = event.Run(sub.StartBlock, sub.StartBlock+1)
		if err != nil {
			log.Fatal("fail to run:", err)
		}
	}
}
