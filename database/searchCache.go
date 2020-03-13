package database

import (
	"time"
	"context"
	"runtime/trace"
	"github.com/patrickmn/go-cache"
)

// SearchCache is a struct for managing a cache of search terms and their post IDs
type SearchCache struct {
	cache *cache.Cache
}

func (c *SearchCache) init() {
	if c.cache == nil {
		c.cache = cache.New(10*time.Second, 5*time.Second)
	}
}

func (c *SearchCache) Get(ctx context.Context, tags string) (p []int64, found bool) {
	defer trace.StartRegion(ctx, "DB/SearchCache/Get").End()
	result, ok := c.cache.Get(tags)
	found = ok
	if found {
		p = result.([]int64)
	}
	return
}

func (c *SearchCache) Add(ctx context.Context, tags string, values []int64) {
	defer trace.StartRegion(ctx, "DB/SearchCache/Add").End()
	c.cache.Set(tags, values, cache.DefaultExpiration)
}

func (c *SearchCache) Init() {
	c.init()
}
