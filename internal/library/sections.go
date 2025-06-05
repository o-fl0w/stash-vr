package library

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"slices"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/filter"
	"stash-vr/internal/stash/gql"
	"strings"
	"sync"
)

type Section struct {
	Name string
	Ids  []string
}

func (service *Service) GetSections(ctx context.Context) ([]Section, error) {
	res, err, _ := service.single.Do("sections", func() (interface{}, error) {
		filters, err := service.getFilters(ctx)
		if err != nil {
			return nil, err
		}

		var sections []Section
		if filters != nil {
			sections, err = service.getSectionsByFilters(ctx, filters)
		} else {
			log.Ctx(ctx).Info().Msg("No saved scene filters found, creating default section with ALL scenes")
			sections, err = service.getDefaultSections(ctx)
		}
		if err != nil {
			return nil, err
		}

		linkCount := 0
		service.mu.Lock()
		for k := range service.vdCache {
			delete(service.vdCache, k)
		}

		for _, v := range sections {
			linkCount += len(v.Ids)
			for _, id := range v.Ids {
				service.vdCache[id] = nil
			}
		}
		service.mu.Unlock()

		log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", linkCount).
			Int("scenes", len(service.vdCache)).
			Msg("Index built")

		return sections, nil
	})
	if err != nil {
		return nil, err
	}
	return res.([]Section), nil
}

func (service *Service) getDefaultSections(ctx context.Context) ([]Section, error) {
	resp, err := gql.FindAllSceneIds(ctx, service.StashClient)
	if err != nil {
		return nil, fmt.Errorf("FindAllSceneIds: %w", err)
	}
	allScenesSection := Section{
		Name: "All",
		Ids:  make([]string, len(resp.FindScenes.Scenes)),
	}
	for i := range resp.FindScenes.Scenes {
		allScenesSection.Ids[i] = resp.FindScenes.Scenes[i].Id
	}
	return []Section{allScenesSection}, nil
}

func (service *Service) getSectionsByFilters(ctx context.Context, filters []gql.SavedFilterParts) ([]Section, error) {
	sections := make([]Section, len(filters))

	wg := sync.WaitGroup{}
	wg.Add(len(filters))

	for i, f := range filters {
		go func(i int, f gql.SavedFilterParts) {
			defer wg.Done()
			flog := log.Ctx(ctx).With().Str("filterId", f.Id).Str("name", f.Name).Logger()

			sceneFilter, err := filter.SavedFilterToSceneFilter(ctx, f)
			if err != nil {
				flog.Warn().Err(err).Interface("savedFilter", f).Msg("Failed to convert filter, skipping")
				return
			}

			resp, err := gql.FindSceneIdsByFilter(ctx, service.StashClient, &sceneFilter.SceneFilter, &sceneFilter.FilterOpts)
			if err != nil {
				flog.Err(err).Interface("savedFilter", f).Interface("sceneFilter", sceneFilter).Msg("Failed to find scenes by filter, skipping")
				return
			}

			if len(resp.FindScenes.Scenes) == 0 {
				flog.Debug().Msg("Filter skipped: 0 scenes")
				return
			}

			sections[i] = Section{
				Name: f.Name,
				Ids:  make([]string, len(resp.FindScenes.Scenes)),
			}
			for j, v := range resp.FindScenes.Scenes {
				sections[i].Ids[j] = v.Id
			}

			flog.Debug().Int("scenes", len(sections[i].Ids)).Msg("Section built")
		}(i, f)
	}
	wg.Wait()
	sections = slices.DeleteFunc(sections, func(s Section) bool {
		return len(s.Ids) == 0
	})
	return sections, nil
}

func (service *Service) getFilters(ctx context.Context) ([]gql.SavedFilterParts, error) {
	savedFilters, err := gql.FindSavedSceneFilters(ctx, service.StashClient)
	if err != nil {
		return nil, fmt.Errorf("failed to find saved filters: %w", err)
	}

	filterByIDs := func(ids []string) []gql.SavedFilterParts {
		var out []gql.SavedFilterParts
		for _, id := range ids {
			for _, f := range savedFilters.FindSavedFilters {
				if f.Id == id {
					out = append(out, f.SavedFilterParts)
					break
				}
			}
		}
		return out
	}

	switch {
	case config.Get().Filters == "frontpage":
		fpIds, err := stash.FindSavedFilterIdsByFrontPage(ctx, service.StashClient)
		if err != nil {
			return nil, fmt.Errorf("failed to find frontpage filters: %w", err)
		}
		return filterByIDs(fpIds), nil

	case config.Get().Filters != "":
		ids := strings.Split(config.Get().Filters, ",")
		for i := range ids {
			ids[i] = strings.Trim(ids[i], " \"'")
		}
		return filterByIDs(ids), nil

	default:
		fpIds, err := stash.FindSavedFilterIdsByFrontPage(ctx, service.StashClient)
		if err != nil {
			return nil, fmt.Errorf("failed to find frontpage filter IDs: %w", err)
		}
		front := filterByIDs(fpIds)

		seen := make(map[string]struct{}, len(fpIds))
		for _, id := range fpIds {
			seen[id] = struct{}{}
		}
		var rest []gql.SavedFilterParts
		for _, f := range savedFilters.FindSavedFilters {
			if _, ok := seen[f.Id]; !ok {
				rest = append(rest, f.SavedFilterParts)
			}
		}
		return append(front, rest...), nil
	}
}
