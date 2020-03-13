package database

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SearchCache is a struct for managing a cache of search terms and their post IDs
type SearchCache struct {
	cache map[string][]int64
	lock  sync.RWMutex
	times map[string]time.Time
}

func (c *SearchCache) init() {
	if c.cache == nil {
		c.cache = make(map[string][]int64)
		c.times = make(map[string]time.Time)
	}
}

func (c *SearchCache) Get(tags string) ([]int64, bool) {
	c.lock.RLock()
	val, ok := c.cache[tags]
	if ok {
		log.Debug().Msg(tags + " fetched from search cache.")
	} else {
		log.Debug().Msg(tags + " not in search cache.")
	}
	c.lock.RUnlock()
	return val, ok
}
func (c *SearchCache) Add(tags string, values []int64) {
	c.lock.Lock()
	if _, ok := c.cache[tags]; !ok {
		c.cache[tags] = values
		c.times[tags] = time.Now()
		log.Debug().Msg(tags + " added to search cache.")
	}
	c.lock.Unlock()
}

func (c *SearchCache) Start() {
	c.init()
	for {
		c.lock.Lock()
		for tags := range c.cache {
			val, ok := c.times[tags]
			if !ok || (time.Now().After(val.Add(time.Second * 10))) {
				log.Debug().Msg(tags + " has expired, removing from search cache.")
				delete(c.cache, tags)
				delete(c.times, tags)
			}
		}
		c.lock.Unlock()
		time.Sleep(time.Second)
	}
}
