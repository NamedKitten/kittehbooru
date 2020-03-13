package database

import (
	"github.com/patrickmn/go-cache"
	"time"
	"context"
	"runtime/trace"
	"fmt"

	"github.com/NamedKitten/kittehimageboard/types"
)

// UserCache is a struct for managing a cache of users
type UserCache struct {
	cache *cache.Cache
}

func (c *UserCache) init() {
	if c.cache == nil {
		c.cache = cache.New(10*time.Second, 5*time.Second)
	}
}

func (c *UserCache) Get(ctx context.Context, username string) (p types.User, ok bool) {
	defer trace.StartRegion(ctx, "DB/UserCache/Get").End()
	var result interface{}
	result, ok = c.cache.Get(username)
	if ok {
		fmt.Println("found!", username)
		p = result.(types.User)
	} else {
		fmt.Println("not found!", username)
	}
	return p, ok
}

func (c *UserCache) Add(ctx context.Context, u types.User) {
	defer trace.StartRegion(ctx, "DB/UserCache/Add").End()
	c.cache.Set(u.Username, u, cache.DefaultExpiration)
	fmt.Println("added!", u.Username)

}

func (c *UserCache) Init() {
	c.init()
}
