package cache

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/section"
	"stash-vr/internal/section/aggregate"
	"sync"
)

var cache sections

type sections struct {
	lock     sync.RWMutex
	sections []section.Section
}

func (c *sections) get() []section.Section {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.sections
}

func (c *sections) set(ss []section.Section) {
	c.lock.Lock()
	c.sections = ss
	c.lock.Unlock()
}

func GetSections(ctx context.Context, client graphql.Client) []section.Section {
	cached := cache.get()
	if len(cached) == 0 {
		log.Ctx(ctx).Trace().Msg("Cache miss")
		c := aggregate.Build(ctx, client)
		cache.set(c)
		return c
	} else {
		log.Ctx(ctx).Trace().Msg("Cache hit")
		go func() {
			ctx = log.Ctx(ctx).With().Str("op", "bg").Logger().WithContext(context.Background())
			log.Ctx(ctx).Trace().Msg("Prefetching...")
			c := aggregate.Build(ctx, client)
			cache.set(c)
		}()
		return cached
	}
}
