package database

import (
	"context"
	"github.com/patrickmn/go-cache"
	"runtime/trace"
	"time"
)

type ContextCache struct {
	c    *cache.Cache
	name string
}

func (c ContextCache) Get(ctx context.Context, k string) (interface{}, bool) {
	defer trace.StartRegion(ctx, c.name+"/Get").End()
	return c.c.Get(k)
}

func (c ContextCache) Add(ctx context.Context, k string, x interface{}, d time.Duration) {
	defer trace.StartRegion(ctx, c.name+"/Add").End()
	c.c.Add(k, x, d)
}

func (c ContextCache) Set(ctx context.Context, k string, x interface{}, d time.Duration) {
	defer trace.StartRegion(ctx, c.name+"/Set").End()
	c.c.Set(k, x, d)
}

var userCache = ContextCache{cache.New(time.Minute, time.Minute), "userCache"}
var postTagsCache = ContextCache{cache.New(5*time.Minute, time.Minute), "postTagsCache"}
var searchCache = ContextCache{cache.New(time.Minute, time.Minute/2), "searchCache"}
var tagCountsCache = ContextCache{cache.New(5*time.Minute, time.Minute), "tagCountsCache"}
var sessionCache = ContextCache{cache.New(time.Minute, time.Minute), "sessionCache"}
