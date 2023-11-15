package cache

import (
	"context"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
	"sync"
)

type Cache[T any] struct {
	dataLock   sync.RWMutex
	fetchMutex sync.Mutex
	data       *T
}

var single = singleflight.Group{}

func (c *Cache[T]) Get(ctx context.Context, fetch func(ctx context.Context) T) T {

	if c.data == nil {
		d, _, _ := single.Do("", func() (interface{}, error) {
			d := fetch(ctx)
			c.data = &d
			return d, nil
		})
		return d.(T)
	}

	go func(ctx context.Context) {
		ctx = log.Ctx(ctx).With().Str("op", "bg").Logger().WithContext(context.Background())
		single.Do("", func() (interface{}, error) {
			log.Ctx(ctx).Trace().Msg("Prefetching...")
			d := fetch(ctx)
			c.data = &d
			return d, nil
		})
	}(ctx)
	return *c.data

}
