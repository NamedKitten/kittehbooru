package database

import (
	"context"
	"github.com/patrickmn/go-cache"
	"runtime/trace"
	"time"

	"github.com/NamedKitten/kittehimageboard/types"
)

// TagCountsCache is a struct for managing a cache of tag counts
type TagCountsCache struct {
	cache *cache.Cache
}

func (c *TagCountsCache) init() {
	if c.cache == nil {
		c.cache = cache.New(10*time.Second, 5*time.Second)
	}
}

func (c *TagCountsCache) Get(ctx context.Context, query string) (tc []types.TagCounts, found bool) {
	defer trace.StartRegion(ctx, "DB/TagCountsCache/Get").End()
	result, ok := c.cache.Get(query)
	found = ok
	if found {
		tc = result.([]types.TagCounts)
	}
	return
}

func (c *TagCountsCache) Add(ctx context.Context, query string, t []types.TagCounts) {
	defer trace.StartRegion(ctx, "DB/TagCountsCache/Add").End()
	c.cache.Set(query, t, cache.DefaultExpiration)
}

func (c *TagCountsCache) Init() {
	c.init()
}
