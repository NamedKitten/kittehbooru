package database

import (
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type SearchCache struct {
	cache map[string][]int64
	lock  sync.Mutex
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
	for true {
		c.lock.Lock()
		for tags := range c.cache {
			val, ok := c.times[tags]
			if !ok || (val <= time.Now().Add(time.Second).Unix()) {
				log.Info().Msg(tags + " has expired, removing from cache.")
				delete(c.cache, tags)
				delete(c.times, tags)
			}
		}
		c.lock.Unlock()

		time.Sleep(time.Second * 2)
	}
}
