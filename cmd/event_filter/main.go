package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	contractevent "github.com/lengzhao/contract_event"
	"gopkg.in/yaml.v2"
)

func main() {
	confFile := flag.String("conf", "./config.yaml", "config file(yaml)")
	flag.Parse()
	data, err := os.ReadFile(*confFile)
	if err != nil {
		log.Fatal("fail to read config file:", *confFile, err)
	}
	var conf contractevent.Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal("fail to unmarshal config:", err)
	}
	mgr, err := contractevent.NewManager(conf)
	if err != nil {
		log.Fatal("fail to NewManager:", err)
	}
	mgr.Run()

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	<-sign
	mgr.Close()
}
