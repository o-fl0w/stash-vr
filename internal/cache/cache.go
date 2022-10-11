package cache

import (
	"context"
	"github.com/rs/zerolog/log"
	"sync"
)

type Cache[T any] struct {
	dataLock   sync.RWMutex
	fetchMutex sync.Mutex
	data       *T
}

func (c *Cache[T]) Get(ctx context.Context, fetch func(ctx context.Context) T) T {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	if c.data == nil {
		d := fetch(ctx)
		c.data = &d
		return *c.data
	} else {
		go func(ctx context.Context) {
			ctx = log.Ctx(ctx).With().Str("op", "bg").Logger().WithContext(context.Background())
			if c.fetchMutex.TryLock() {
				defer c.fetchMutex.Unlock()
				log.Ctx(ctx).Trace().Msg("Prefetching...")
				d := fetch(ctx)
				c.dataLock.Lock()
				defer c.dataLock.Unlock()
				c.data = &d
			} else {
				log.Ctx(ctx).Trace().Msg("Already fetching...")
			}
		}(ctx)
		return *c.data
	}
}
