package common

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common/cache"
	"stash-vr/internal/api/common/types"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/filter/scenefilter"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
	"sync"
)

func RefreshCache(ctx context.Context, client graphql.Client) []types.Section {
	ctx = log.With().Logger().WithContext(ctx)
	c := buildIndex(ctx, client)
	cache.Store.Index.Set(c)
	return c
}

func GetIndex(ctx context.Context, client graphql.Client) []types.Section {
	cached := cache.Store.Index.Get()
	if len(cached) == 0 {
		return RefreshCache(ctx, client)
	} else {
		go func() {
			c := buildIndex(ctx, client)
			cache.Store.Index.Set(c)
		}()
		return cached
	}
}

func buildIndex(ctx context.Context, client graphql.Client) []types.Section {
	var sss [2][]types.Section

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		sss[0], err = sectionsByFrontPage(ctx, client, "")
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by front page")
			return
		}
	}()

	if !config.Get().FrontPageFiltersOnly {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			sss[1], err = sectionsBySavedFilters(ctx, client, "?:")
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by saved filters")
				return
			}
		}()
	}

	wg.Wait()

	var sections []types.Section

	for _, ss := range sss {
		for _, s := range ss {
			if s.FilterId != "" && containsSavedFilterId(s.FilterId, sections) {
				log.Ctx(ctx).Debug().Str("filterId", s.FilterId).Str("section", s.Name).Msg("Filter already added, skipping")
				continue
			}
			sections = append(sections, s)
		}
	}

	var links int
	for _, section := range sections {
		links += len(section.PreviewPartsList)
	}

	if links > 10000 {
		log.Ctx(ctx).Warn().Int("links", links).Msg("More than 10.000 links generated. Known to cause issues with video players.")
	}

	log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", links).Msg("Index built")

	return sections
}

func sectionsByFrontPage(ctx context.Context, client graphql.Client, prefix string) ([]types.Section, error) {
	savedFilters, err := stash.FindSavedSceneFiltersByFrontPage(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFiltersByFrontPage: %w", err)
	}

	sections := util.Transformation[gql.SavedFilterParts, types.Section]{
		Transform: func(savedFilter gql.SavedFilterParts) (types.Section, error) {
			return sectionFromSavedSceneFilter(ctx, client, prefix, savedFilter)
		},
		Success: util.Ptr(func(savedFilter gql.SavedFilterParts, section types.Section) {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Str("section", section.Name).Int("links", len(section.PreviewPartsList)).Msg("Section built from Front Page")
		}),
		Failure: util.Ptr(func(savedFilter gql.SavedFilterParts, e error) {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Skipped filter: sectionsByFrontPage: sectionFromSavedSceneFilter")
		}),
	}.Ordered(savedFilters)

	return sections, nil
}

func sectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string) ([]types.Section, error) {
	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	savedFilters := make([]gql.SavedFilterParts, len(savedFiltersResponse.FindSavedFilters))
	for i, s := range savedFiltersResponse.FindSavedFilters {
		savedFilters[i] = s.SavedFilterParts
	}

	sections := util.Transformation[gql.SavedFilterParts, types.Section]{
		Transform: func(savedFilter gql.SavedFilterParts) (types.Section, error) {
			return sectionFromSavedSceneFilter(ctx, client, prefix, savedFilter)
		},
		Success: util.Ptr(func(savedFilter gql.SavedFilterParts, section types.Section) {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Str("section", savedFilter.Name).Int("links", len(section.PreviewPartsList)).Msg("Section built from Saved Filter")
		}),
		Failure: util.Ptr(func(savedFilter gql.SavedFilterParts, e error) {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Skipped filter: sectionsBySavedFilters: sectionFromSavedSceneFilter")
		}),
	}.Ordered(savedFilters)

	return sections, nil
}

func sectionFromSavedSceneFilter(ctx context.Context, client graphql.Client, prefix string, savedFilter gql.SavedFilterParts) (types.Section, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return types.Section{}, fmt.Errorf("not a scene filter, mode='%s'", savedFilter.Mode)
	}

	filterQuery, err := scenefilter.ParseJsonEncodedFilter(savedFilter.Filter)
	if err != nil {
		return types.Section{}, fmt.Errorf("ParseJsonEncodedFilter: %w", err)
	}

	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filterQuery.SceneFilter, &filterQuery.FilterOpts)
	if err != nil {
		return types.Section{}, fmt.Errorf("FindScenesByFilter savedFilter=%+v parsedFilter=%+v: %w", savedFilter, util.AsJsonStr(filterQuery), err)
	}

	if len(scenesResponse.FindScenes.Scenes) == 0 {
		return types.Section{}, fmt.Errorf("0 videos found")
	}

	section := types.Section{
		Name:             getFilterName(prefix, savedFilter),
		FilterId:         savedFilter.Id,
		PreviewPartsList: make([]gql.ScenePreviewParts, 0, len(scenesResponse.FindScenes.Scenes)),
	}

	for _, s := range scenesResponse.FindScenes.Scenes {
		section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
	}
	return section, nil
}

func containsSavedFilterId(id string, list []types.Section) bool {
	for _, v := range list {
		if id == v.FilterId {
			return true
		}
	}
	return false
}

func getFilterName(prefix string, fp gql.SavedFilterParts) string {
	var sb strings.Builder
	sb.WriteString(prefix)
	if fp.Name == "" {
		sb.WriteString("<")
		sb.WriteString(fp.Id)
		sb.WriteString(">")
	} else {
		sb.WriteString(fp.Name)
	}
	return sb.String()
}
