package library

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"slices"
	"stash-vr/internal/config"
	"stash-vr/internal/stash/filter"
	"stash-vr/internal/stash/gql"
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

		service.muVdCache.Lock()
		for k := range service.vdCache {
			delete(service.vdCache, k)
		}

		service.Stats.Links = 0
		for _, v := range sections {
			service.Stats.Links += len(v.Ids)
			for _, id := range v.Ids {
				service.vdCache[id] = nil
			}
		}
		service.Stats.Scenes = len(service.vdCache)
		service.muVdCache.Unlock()

		log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", service.Stats.Links).
			Int("scenes", service.Stats.Scenes).
			Msg("Index built")

		_ = service.LoadTags(ctx)

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

	userConfigFilters := config.User(ctx).Filters
	out := buildActiveFilters(savedFilters.FindSavedFilters, userConfigFilters)
	return out, nil
}

func buildActiveFilters(stashFilters []*gql.FindSavedSceneFiltersFindSavedFiltersSavedFilter, cfgFilters []config.Filter) []gql.SavedFilterParts {
	if len(cfgFilters) == 0 {
		out := make([]gql.SavedFilterParts, 0, len(stashFilters))
		for _, s := range stashFilters {
			out = append(out, s.SavedFilterParts)
		}
		return out
	}

	stashFilterParts := make(map[string]gql.SavedFilterParts, len(stashFilters))
	for _, sf := range stashFilters {
		stashFilterParts[sf.Id] = sf.SavedFilterParts
	}

	out := make([]gql.SavedFilterParts, 0, len(stashFilters))
	seen := make(map[string]struct{}, len(stashFilters))

	// 1) Enabled cfgFilters in the given order.
	for _, cf := range cfgFilters {
		seen[cf.ID] = struct{}{}
		if cf.Disabled {
			continue
		}
		sf, ok := stashFilterParts[cf.ID]
		if !ok {
			continue
		}

		if cf.Name != "" {
			sf.Name = cf.Name
		}
		out = append(out, sf)
	}

	for _, s := range stashFilters {
		if _, done := seen[s.Id]; done {
			continue
		}
		out = append(out, s.SavedFilterParts)
	}

	return out
}
