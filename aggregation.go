package contractevent

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Manager struct {
	conf         Config
	db           *gorm.DB
	events       map[string]*Event
	notification map[string]*NotifyTask
	chain        *chain
	router       *gin.Engine
	wg           sync.WaitGroup
	stopping     chan int
}

func NewManager(conf Config) (*Manager, error) {
	var out Manager
	out.conf = conf
	out.stopping = make(chan int)
	chain, err := newChain(conf.Chain.RPCNode, conf.Chain.DelayBlock)
	if err != nil {
		return nil, err
	}
	out.chain = chain
	db, err := NewDB(conf.DB)
	if err != nil {
		return nil, err
	}
	CreateBlockRecord(db)
	CreateNotifyRecord(db)
	out.db = db
	out.events = make(map[string]*Event)
	out.notification = make(map[string]*NotifyTask)
	for _, it := range conf.Subs {
		if _, ok := out.events[it.Alias]; ok {
			log.Error("exist alias:", it.Alias)
			return nil, fmt.Errorf("exist alias:%s", it.Alias)
		}
		event, err := NewEventWithDB(it, chain.client, db)
		if err != nil {
			log.Error("fail to new event:", it.Alias, err)
			return nil, err
		}
		out.events[it.Alias] = event
		if it.WebHook != "" {
			out.notification[it.Alias] = NewNotifyTask(db, it.Alias, it.WebHook)
		}
	}

	if conf.Http.Port > 0 {
		eng := gin.Default()
		group := eng.Group(conf.Http.PrefixPath)
		HttpRouter(group, db)
		out.router = eng
	}

	return &out, nil
}

func (m *Manager) Run() {
	if m.conf.Http.Port > 0 {
		go func() {
			err := m.router.Run(fmt.Sprintf(":%d", m.conf.Http.Port))
			if err != nil {
				log.Error("http server error:", m.conf.Http.Port, err)
			}
		}()
	}
	for alias, it := range m.events {
		m.wg.Add(1)
		go func(alias string, event *Event, c *chain) {
			var wTime time.Duration = 100
			for {
				select {
				case <-time.After(time.Millisecond * wTime):
					wTime = 100
				case <-m.stopping:
					m.wg.Done()
					return
				}
				last := c.SafeBlockNumber()
				bn, err := GetBlockRecord(m.db, alias)
				if err != nil {
					log.Error("fail to get block record:", alias, err)
					wTime = 10000
					continue
				}
				if bn >= last {
					wTime = 2000
					continue
				}
				bn++
				if last > bn+5 {
					last = bn + 5
				}
				err = event.Run(bn, last)
				if err != nil {
					log.Error("fail to event.Run:", alias, err)
					wTime = 5000
					continue
				}
				SetBlockRecord(m.db, alias, last)
			}
		}(alias, it, m.chain)
	}

	for alias, it := range m.notification {
		m.wg.Add(1)
		go func(alias string, ntf *NotifyTask) {
			var wTime time.Duration = 100
			for {
				select {
				case <-time.After(time.Millisecond * wTime):
					wTime = 1000
				case <-m.stopping:
					m.wg.Done()
					return
				}
				err := ntf.Run(10)
				if err != nil {
					wTime = 3000
				}
			}
		}(alias, it)
	}
	m.wg.Wait()
}

func (m *Manager) Close() {
	for i := 0; i < len(m.events)+len(m.notification); i++ {
		m.stopping <- 1
	}
}
