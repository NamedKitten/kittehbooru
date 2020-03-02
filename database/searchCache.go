package database

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type SearchCache struct {
	cache map[string][]int64
	lock  sync.RWMutex
	times map[string]int64
}

func (c *SearchCache) init() {
	if c.cache == nil {
		c.cache = make(map[string][]int64)
		c.times = make(map[string]int64)
	}
}

func (c *SearchCache) Get(tags string) ([]int64, bool) {
	c.lock.Lock()
	val, ok := c.cache[tags]
	c.lock.Unlock()
	return val, ok
}
func (c *SearchCache) Add(tags string, values []int64) {
	c.lock.Lock()
	c.cache[tags] = values
	c.times[tags] = time.Now().Unix()
	c.lock.Unlock()
}

func (c *SearchCache) Start() {
	c.init()
	for {
		c.lock.Lock()
		for tags := range c.cache {
			val, ok := c.times[tags]
			if !ok || (val <= time.Now().Add(time.Minute).Unix()) {
				log.Info().Msg(tags + " has expired, removing from cache.")
				delete(c.cache, tags)
				delete(c.times, tags)
			}
		}
		c.lock.Unlock()

		time.Sleep(time.Second * 2)
	}
}
