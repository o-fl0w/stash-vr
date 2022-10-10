package sections

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/sections/section"
	"sync"
)

var dataLock sync.RWMutex
var fetchMutex sync.Mutex
var cSections []section.Section

func Get(ctx context.Context, client graphql.Client) []section.Section {
	dataLock.Lock()
	defer dataLock.Unlock()
	if cSections == nil {
		log.Ctx(ctx).Trace().Msg("Cache miss")
		cSections = build(ctx, client)
		return cSections
	} else {
		log.Ctx(ctx).Trace().Msg("Cache hit")
		go func(ctx context.Context) {
			ctx = log.Ctx(ctx).With().Str("op", "bg").Logger().WithContext(context.Background())
			if fetchMutex.TryLock() {
				defer fetchMutex.Unlock()
				log.Ctx(ctx).Trace().Msg("Prefetching...")
				ss := build(ctx, client)
				dataLock.Lock()
				defer dataLock.Unlock()
				cSections = ss
			} else {
				log.Ctx(ctx).Trace().Msg("Already fetching...")
			}
		}(ctx)
		return cSections
	}
}
