package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	contractevent "github.com/lengzhao/contract_event"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	contractevent.Config `yaml:",inline"`
	LogFile              string `yaml:"log_file,omitempty"`
	LogLevel             uint32 `yaml:"log_level,omitempty"`
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
	if conf.LogFile != "" {
		fn, err := os.Create(conf.LogFile)
		if err != nil {
			log.Fatal("fail to create log file:", conf.LogFile, err)
		}
		defer fn.Close()
		log.SetOutput(fn)
	}
	if conf.LogLevel > 0 {
		log.SetLevel(log.Level(conf.LogLevel))
	}
	mgr, err := contractevent.NewManager(conf.Config)
	if err != nil {
		log.Fatal("fail to NewManager:", err)
	}
	go mgr.Run()

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	<-sign
	log.Warn("receive signal to exit")
	mgr.Close()
}
