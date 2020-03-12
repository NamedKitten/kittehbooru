package database

import (
	"sync"
	"time"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

// TagCountsCache is a struct for managing a cache of tag counts
type TagCountsCache struct {
	cache map[string][]types.TagCounts
	lock  sync.RWMutex
}

func (c *TagCountsCache) init() {
	if c.cache == nil {
		c.cache = make(map[string][]types.TagCounts)
	}
}

func (c *TagCountsCache) Get(tags string) ([]types.TagCounts, bool) {
	c.lock.Lock()
	val, ok := c.cache[tags]
	c.lock.Unlock()
	if ok {
		log.Debug().Msg(tags + " fetched from tag counts cache.")
	} else {
		log.Debug().Msg(tags + " not in tag counts cache.")
	}
	return val, ok
}
func (c *TagCountsCache) Add(tags string, values []types.TagCounts) {
	c.lock.Lock()
	c.cache[tags] = values
	c.lock.Unlock()
	log.Debug().Msg(tags + " added to tag counts cache.")
}

func (c *TagCountsCache) Start() {
	c.init()
	for {
		c.lock.Lock()
		for tags := range c.cache {
			log.Debug().Msg(tags + " has expired, removing from tag counts cache.")
			delete(c.cache, tags)
		}
		c.lock.Unlock()
		time.Sleep(time.Minute)
	}
}
