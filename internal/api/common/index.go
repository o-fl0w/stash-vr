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
)

type Section struct {
	Name             string
	FilterId         string
	PreviewPartsList []gql.ScenePreviewParts
}

func BuildIndex(ctx context.Context, client graphql.Client) []Section {
	var sections []Section

	if err := sectionsDefault(ctx, client, "", &sections); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to build default sections")
	}

	if err := sectionsByFrontPage(ctx, client, "", &sections); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by front page")
	}

	if err := sectionsBySavedFilters(ctx, client, "?:", &sections); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by saved filters")
	}

	//if err := sectionsByTags(ctx, client, baseUrl, "#:", &index.Scenes, &sections); err != nil {
	//	return Index{}, fmt.Errorf("sectionsByTags: %w", err)
	//}

	log.Ctx(ctx).Info().Int("sectionCount", len(sections)).Msg("Index built")

	return sections
}

func sectionsDefault(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	section := Section{Name: fmt.Sprintf("%s%s", prefix, "All")}

	scenesResponse, err := gql.FindAllScenes(ctx, client)
	if err != nil {
		return fmt.Errorf("FindAllScenes: %w", err)
	}
	for _, s := range scenesResponse.FindScenes.Scenes {
		section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
	}
	*destination = append(*destination, section)

	return nil
}

func sectionsByFrontPage(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	savedSceneFilters, err := stash.FindSavedSceneFiltersByFrontPage(ctx, client)
	if err != nil {
		return fmt.Errorf("FindSavedSceneFiltersByFrontPage: %w", err)
	}

	for _, savedFilter := range savedSceneFilters {
		section, err := sectionFromSavedSceneFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Skipped filter: sectionsByFrontPage: sectionFromSavedSceneFilter")
			continue
		}
		*destination = append(*destination, section)

		log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Int("videoCount", len(section.PreviewPartsList)).Msg("Section added from Front Page")
	}

	return nil
}

func sectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	for _, savedFilter := range savedFiltersResponse.FindSavedFilters {
		if containsSavedFilterId(savedFilter.Id, *destination) {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Filter already added, skipping")
			continue
		}

		section, err := sectionFromSavedSceneFilter(ctx, client, prefix, savedFilter.SavedFilterParts)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Skipped filter: sectionsBySavedFilters: sectionFromSavedSceneFilter")
			continue
		}
		*destination = append(*destination, section)

		log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Int("videoCount", len(section.PreviewPartsList)).Msg("Section added from Saved Filter")
	}

	return nil
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
		Name:     getFilterName(prefix, savedFilter),
		FilterId: savedFilter.Id,
	}
	for _, s := range scenesResponse.FindScenes.Scenes {
		section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
	}
	return section, nil
}

func sectionsByTags(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	findTagsResponse, err := gql.FindAllNonEmptyTags(ctx, client)
	if err != nil {
		return fmt.Errorf("FindAllNonEmptyTags: %w", err)
	}

	var tagIds []string
	for _, tag := range findTagsResponse.FindTags.Tags {
		tagIds = append(tagIds, tag.Id)
	}

	scenesResponse, err := gql.FindScenesByTags(ctx, client, tagIds)
	if err != nil {
		return fmt.Errorf("FindScenesByTags: %w", err)
	}

	tagMap := make(map[string]Section)
	for _, s := range scenesResponse.FindScenes.Scenes {
		for _, tag := range s.Tags {
			name := fmt.Sprintf("%s%s", prefix, tag.Name)
			if v, ok := tagMap[name]; !ok {
				tagMap[name] = Section{
					Name:             name,
					FilterId:         "",
					PreviewPartsList: []gql.ScenePreviewParts{s.ScenePreviewParts},
				}
			} else {
				v.PreviewPartsList = append(v.PreviewPartsList, s.ScenePreviewParts)
			}
		}
	}

	for _, v := range tagMap {
		*destination = append(*destination, v)
	}

	return nil
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
