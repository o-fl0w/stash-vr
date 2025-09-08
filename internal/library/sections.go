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
	"sync"
)

type Section struct {
	Name string
	Ids  []string
}

func (libraryService *Service) GetSections(ctx context.Context) ([]Section, error) {
	res, err, _ := libraryService.single.Do("sections", func() (interface{}, error) {
		filters, err := libraryService.getFilters(ctx)
		if err != nil {
			return nil, err
		}

		var sections []Section
		if len(filters) == 0 {
			log.Ctx(ctx).Info().Msg("No saved scene filters found, creating default section with ALL scenes")
			sections, err = libraryService.getDefaultSections(ctx)
		} else {
			sections, err = libraryService.getSectionsByFilters(ctx, filters)
		}
		if err != nil {
			return nil, err
		}

		libraryService.muVdCache.Lock()
		for k := range libraryService.vdCache {
			delete(libraryService.vdCache, k)
		}

		libraryService.Stats.Links = 0
		for _, v := range sections {
			libraryService.Stats.Links += len(v.Ids)
			for _, id := range v.Ids {
				libraryService.vdCache[id] = nil
			}
		}
		libraryService.Stats.Scenes = len(libraryService.vdCache)
		libraryService.muVdCache.Unlock()

		log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", libraryService.Stats.Links).
			Int("scenes", libraryService.Stats.Scenes).
			Msg("Index built")

		_ = libraryService.LoadTags(ctx)

		log.Ctx(ctx).Debug().Int("tags", len(libraryService.tagCache)).Msg("Cached tags")

		return sections, nil
	})
	if err != nil {
		return nil, err
	}
	return res.([]Section), nil
}

func (libraryService *Service) getDefaultSections(ctx context.Context) ([]Section, error) {
	resp, err := gql.FindAllSceneIds(ctx, libraryService.StashClient)
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

func (libraryService *Service) getSectionsByFilters(ctx context.Context, filters []gql.SavedFilterParts) ([]Section, error) {
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

			resp, err := gql.FindSceneIdsByFilter(ctx, libraryService.StashClient, &sceneFilter.SceneFilter, &sceneFilter.FilterOpts)
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

func (libraryService *Service) getFilters(ctx context.Context) ([]gql.SavedFilterParts, error) {
	savedFilters, err := gql.FindSavedSceneFilters(ctx, libraryService.StashClient)
	if err != nil {
		return nil, fmt.Errorf("failed to find saved filters: %w", err)
	}

	if len(savedFilters.FindSavedFilters) == 0 {
		return nil, nil
	}

	var out []gql.SavedFilterParts

	userConfigFilters := config.User(ctx).Filters

	if len(userConfigFilters) == 0 {
		out, err = libraryService.buildFiltersByFrontpage(ctx, savedFilters)
	} else {
		out = buildFiltersByUserConfig(ctx, savedFilters, userConfigFilters)
	}
	return out, nil
}

func (libraryService *Service) buildFiltersByFrontpage(ctx context.Context, savedFilters *gql.FindSavedSceneFiltersResponse) ([]gql.SavedFilterParts, error) {
	fpIds, err := stash.FindSavedFilterIdsByFrontPage(ctx, libraryService.StashClient)
	if err != nil {
		return nil, fmt.Errorf("failed to find frontpage filter IDs: %w", err)
	}

	var front []gql.SavedFilterParts

	for _, id := range fpIds {
		for _, f := range savedFilters.FindSavedFilters {
			if f.Id == id {
				front = append(front, f.SavedFilterParts)
				break
			}
		}
	}

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
	out := append(front, rest...)

	log.Ctx(ctx).Debug().Int("count", len(out)).Msg("Filters built by frontpage")

	return out, nil
}

func buildFiltersByUserConfig(ctx context.Context, savedFilters *gql.FindSavedSceneFiltersResponse, cfgFilters []config.Filter) []gql.SavedFilterParts {
	stashFilters := savedFilters.FindSavedFilters
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

	log.Ctx(ctx).Debug().Int("count", len(out)).Msg("Filters built by user config")

	return out
}
