package aggregate

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/config"
	"stash-vr/internal/section"
	"stash-vr/internal/section/internal"
	"strings"
	"sync"
)

func Build(ctx context.Context, client graphql.Client) []section.Section {
	sss := make([][]section.Section, 3)

	filters := config.Get().Filters

	wg := sync.WaitGroup{}

	if filters == "frontpage" || filters == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss, err := internal.SectionsByFrontpage(ctx, client, "")
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
			ss, err := internal.SectionsBySavedFilters(ctx, client, "?:")
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
			ss, err := internal.SectionsByFilterIds(ctx, client, "?:", filterIds)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by filter ids")
				return
			}
			sss[2] = ss
			log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from filter list")
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

	if len(sections) == 0 {
		log.Ctx(ctx).Info().Msg("No scenes found using current filters. Adding a default 'All' section.")
		s, err := internal.SectionWithAllScenes(ctx, client)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to build custom section 'All'")
		} else {
			if len(s.PreviewPartsList) == 0 {
				log.Ctx(ctx).Info().Msg("No scenes found in Stash.")
			} else {
				sections = append(sections, s)
			}
		}
	}

	count := section.Count(sections)

	if count.Links > 10000 {
		log.Ctx(ctx).Warn().Int("links", count.Links).Msg("More than 10.000 links generated. Known to cause issues with video players.")
	}

	log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", count.Links).Int("scenes", count.Scenes).Msg("Sections build complete")

	return sections
}
