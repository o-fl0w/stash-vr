package sections

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/cache"
	"stash-vr/internal/config"
	"stash-vr/internal/sections/internal"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stimhub"
	"strings"
	"sync"
)

var c cache.Cache[[]section.Section]

func Get(ctx context.Context, stashClient graphql.Client, stimhubClient *stimhub.Client) []section.Section {
	return c.Get(ctx, func(ctx context.Context) []section.Section {
		sections := build(ctx, stashClient, stimhubClient, config.Get().Filters)

		go func() {
			count := Count(sections)
			if count.Links > 10000 {
				log.Ctx(ctx).Warn().Int("links", count.Links).Msg("More than 10.000 links generated. Known to cause issues with video players.")
			}

			log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", count.Links).Int("scenes", count.Scenes).Msg("Sections built")
		}()

		return sections
	})
}

func build(ctx context.Context, stashClient graphql.Client, stimhubClient *stimhub.Client, filters string) []section.Section {
	sss := make([][]section.Section, 4)

	wg := sync.WaitGroup{}

	if filters == "frontpage" || filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsByFrontpage(ctx, stashClient, "")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by front page")
				return
			}
			sss[0] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from front page")
		}()
	}

	if filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsBySavedFilters(ctx, stashClient, "?:")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by saved filters")
				return
			}
			sss[1] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from saved filters")
		}()
	}

	if filters != "frontpage" && filters != "" {
		filterIds := strings.Split(filters, ",")
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsByFilterIds(ctx, stashClient, "?:", filterIds)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by filter ids")
				return
			}
			sss[2] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from filter list")
		}()
	}

	if stimhubClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := internal.StimSection(ctx, *stimhubClient, stashClient)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by stim server")
				return
			}
			ss := []section.Section{s}
			sss[3] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from stim server")
		}()

	}

	wg.Wait()

	var sections []section.Section

	for _, ss := range sss {
	outer:
		for _, s := range ss {
			if s.FilterId == "" {
				continue
			}
			for _, x := range sections {
				if s.FilterId == x.FilterId {
					log.Ctx(ctx).Trace().Str("filterId", s.FilterId).Str("section", s.Name).Msg("Filter already added, skipping")
					continue outer
				}
			}

			sections = append(sections, s)
		}
	}

	if len(sections) == 0 {
		log.Ctx(ctx).Info().Msg("No scenes found using current filters. Adding a default section with all scenes.")
		s, err := internal.SectionWithAllScenes(ctx, stashClient)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to build default section with all scenes")
		} else {
			if len(s.Scenes) == 0 {
				log.Ctx(ctx).Info().Msg("No scenes found.")
			} else {
				sections = append(sections, s)
			}
		}
	}

	return sections
}
