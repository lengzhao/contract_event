package contractevent

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

type chain struct {
	client      *ethclient.Client
	lastBlock   uint64
	lastSync    time.Time
	delayNumber uint64
	mu          sync.Mutex
}

func newChain(url string, delay uint64) (*chain, error) {
	var out chain
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Errorln("fail to dial eth node:", url, err)
		return nil, err
	}
	out.client = client
	out.lastBlock, err = client.BlockNumber(context.Background())
	if err != nil {
		log.Errorln("fail to get block number:", err)
		return nil, err
	}
	log.Infoln("new block:", out.lastBlock)
	out.lastSync = time.Now()
	out.delayNumber = delay
	return &out, nil
}

func (c *chain) SafeBlockNumber() uint64 {
	if time.Since(c.lastSync) > 2*time.Second {
		c.mu.Lock()
		defer c.mu.Unlock()
		bn, err := c.client.BlockNumber(context.Background())
		c.lastSync = time.Now()
		if err != nil {
			log.Warnln("fail to get block number:", err)
		} else {
			c.lastBlock = bn
			log.Infoln("new block:", c.lastBlock)
		}
	}
	if c.lastBlock > c.delayNumber {
		return c.lastBlock - c.delayNumber
	}
	return 0
}

func (c *chain) Close() {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
}
