package common

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common/cache"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/config"
	"strings"
	"sync"
)

func GetIndex(ctx context.Context, client graphql.Client) []section.Section {
	cached := cache.Store.Index.Get()
	if len(cached) == 0 {
		log.Ctx(ctx).Trace().Msg("Cache miss")
		c := buildIndex(ctx, client)
		cache.Store.Index.Set(c)
		return c
	} else {
		log.Ctx(ctx).Trace().Msg("Cache hit")
		go func() {
			ctx = log.Ctx(ctx).With().Str("op", "bg").Logger().WithContext(context.Background())
			log.Ctx(ctx).Trace().Msg("Prefetching...")
			c := buildIndex(ctx, client)
			cache.Store.Index.Set(c)
		}()
		return cached
	}
}

func buildIndex(ctx context.Context, client graphql.Client) []section.Section {
	sss := make([][]section.Section, 3)

	filters := config.Get().Filters

	wg := sync.WaitGroup{}

	if filters == "frontpage" || filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := section.SectionsByFrontPage(ctx, client, "")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by front page")
				return
			}
			sss[0] = ss
		}()
	}

	if filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := section.SectionsBySavedFilters(ctx, client, "?:")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by saved filters")
				return
			}
			sss[1] = ss
		}()
	}

	if filters != "frontpage" && filters != "" {
		filterIds := strings.Split(filters, ",")
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := section.SectionsByFilterIds(ctx, client, "?:", filterIds)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by filter ids")
				return
			}
			sss[2] = ss
		}()
	}

	wg.Wait()

	var sections []section.Section

	for _, ss := range sss {
		for _, s := range ss {
			if s.FilterId != "" && section.ContainsFilterId(s.FilterId, sections) {
				log.Ctx(ctx).Trace().Str("filterId", s.FilterId).Str("section", s.Name).Msg("Filter already added, skipping")
				continue
			}
			sections = append(sections, s)
		}
	}

	var links int
	for _, s := range sections {
		links += len(s.PreviewPartsList)
	}

	if links > 10000 {
		log.Ctx(ctx).Warn().Int("links", links).Msg("More than 10.000 links generated. Known to cause issues with video players.")
	}

	log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", links).Msg("Index built")

	return sections
}
