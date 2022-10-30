package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	contractevent "github.com/lengzhao/contract_event"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RCPNode       string                           `yaml:"rpc_node"`
	WebHook       string                           `yaml:"web_hook"`
	Subscriptions []contractevent.SubscriptionConf `yaml:"subscriptions"`
}

func main() {
	confFile := flag.String("conf", "config.yaml", "config file(yaml)")
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
		log.Println("fail to do ethclient.Dial:", err)
		os.Exit(2)
	}
	for _, sub := range conf.Subscriptions {
		event, err := contractevent.NewEvent(sub, client, func(alias string, info map[string]interface{}) error {
			data, _ := json.Marshal(info)
			resp, err := http.DefaultClient.Post(conf.WebHook, "application/json", bytes.NewReader(data))
			if err != nil {
				log.Println("fail to notify:", conf.WebHook, alias, err)
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Println("error response code:", alias, resp.StatusCode, resp.Status)
				return fmt.Errorf("fail to notify:%s,http code:%d", conf.WebHook, resp.StatusCode)
			}
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
