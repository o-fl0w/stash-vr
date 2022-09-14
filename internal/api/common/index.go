package common

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/filter/scenefilter"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
	"sync"
)

type Section struct {
	Name             string
	FilterId         string
	PreviewPartsList []gql.ScenePreviewParts
}

func BuildIndex(ctx context.Context, client graphql.Client) []Section {
	var sss [2][]Section

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

	wg.Wait()

	var sections []Section

	for _, ss := range sss {
		for _, s := range ss {
			if s.FilterId != "" && containsSavedFilterId(s.FilterId, sections) {
				log.Ctx(ctx).Debug().Str("filterId", s.FilterId).Str("section", s.Name).Msg("Filter already added, skipping")
				continue
			}
			sections = append(sections, s)
		}
	}

	var videoCount int
	for _, section := range sections {
		videoCount += len(section.PreviewPartsList)
	}

	log.Ctx(ctx).Info().Int("sectionCount", len(sections)).Int("videoDataCount", videoCount).Msg("Index built")

	return sections
}

func sectionsByFrontPage(ctx context.Context, client graphql.Client, prefix string) ([]Section, error) {
	savedFilters, err := stash.FindSavedSceneFiltersByFrontPage(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFiltersByFrontPage: %w", err)
	}

	sections := util.Transformation[gql.SavedFilterParts, Section]{
		Transform: func(savedFilter gql.SavedFilterParts) (Section, error) {
			return sectionFromSavedSceneFilter(ctx, client, prefix, savedFilter)
		},
		Success: util.Ptr(func(savedFilter gql.SavedFilterParts, section Section) {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Str("section", section.Name).Int("videoCount", len(section.PreviewPartsList)).Msg("Section built from Front Page")
		}),
		Failure: util.Ptr(func(savedFilter gql.SavedFilterParts, e error) {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Skipped filter: sectionsByFrontPage: sectionFromSavedSceneFilter")
		}),
	}.Ordered(savedFilters)

	return sections, nil
}

func sectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string) ([]Section, error) {
	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	savedFilters := make([]gql.SavedFilterParts, len(savedFiltersResponse.FindSavedFilters))
	for i, s := range savedFiltersResponse.FindSavedFilters {
		savedFilters[i] = s.SavedFilterParts
	}

	sections := util.Transformation[gql.SavedFilterParts, Section]{
		Transform: func(savedFilter gql.SavedFilterParts) (Section, error) {
			return sectionFromSavedSceneFilter(ctx, client, prefix, savedFilter)
		},
		Success: util.Ptr(func(savedFilter gql.SavedFilterParts, section Section) {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Str("section", savedFilter.Name).Int("videoCount", len(section.PreviewPartsList)).Msg("Section built from Saved Filter")
		}),
		Failure: util.Ptr(func(savedFilter gql.SavedFilterParts, e error) {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Skipped filter: sectionsBySavedFilters: sectionFromSavedSceneFilter")
		}),
	}.Ordered(savedFilters)

	return sections, nil
}

func sectionFromSavedSceneFilter(ctx context.Context, client graphql.Client, prefix string, savedFilter gql.SavedFilterParts) (Section, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return Section{}, fmt.Errorf("not a scene filter, mode='%s'", savedFilter.Mode)
	}

	filterQuery, err := scenefilter.ParseJsonEncodedFilter(savedFilter.Filter)
	if err != nil {
		return Section{}, fmt.Errorf("ParseJsonEncodedFilter: %w", err)
	}

	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filterQuery.SceneFilter, &filterQuery.FilterOpts)
	if err != nil {
		return Section{}, fmt.Errorf("FindScenesByFilter savedFilter=%+v parsedFilter=%+v: %w", savedFilter, util.AsJsonStr(filterQuery), err)
	}

	if len(scenesResponse.FindScenes.Scenes) == 0 {
		return Section{}, fmt.Errorf("0 videos found")
	}

	section := Section{
		Name:             getFilterName(prefix, savedFilter),
		FilterId:         savedFilter.Id,
		PreviewPartsList: make([]gql.ScenePreviewParts, 0, len(scenesResponse.FindScenes.Scenes)),
	}

	for _, s := range scenesResponse.FindScenes.Scenes {
		section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
	}
	return section, nil
}

func containsSavedFilterId(id string, list []Section) bool {
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
