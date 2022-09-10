package common

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash"
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

	if err := sectionsCustom(ctx, client, "", &sections); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("sectionsCustom")
	}

	if err := sectionsByFrontPage(ctx, client, "", &sections); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("sectionsByFrontPage")
	}

	if err := sectionsBySavedFilters(ctx, client, "?:", &sections); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("sectionsBySavedFilters")
	}

	//if err := sectionsByTags(ctx, client, baseUrl, "#:", &index.Scenes, &sections); err != nil {
	//	return Index{}, fmt.Errorf("sectionsByTags: %w", err)
	//}

	log.Ctx(ctx).Info().Int("sectionCount", len(sections)).Msg("Index built")

	return sections
}

func sectionsCustom(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	section := Section{Name: fmt.Sprintf("%s%s", prefix, "All")}
	var sceneFilter gql.SceneFilterType
	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &sceneFilter, "", gql.SortDirectionEnumAsc)
	if err != nil {
		return fmt.Errorf("FindScenesByFilter: %w", err)
	}
	for _, s := range scenesResponse.FindScenes.Scenes {
		section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
	}
	*destination = append(*destination, section)

	return nil
}

func sectionsByFrontPage(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	ids, err := stash.FindFrontPageSavedFilterIds(ctx, client)
	if err != nil {
		return fmt.Errorf("FindFrontPageSavedFilterIds: %w", err)
	}

	for _, id := range ids {
		savedFilterResponse, err := gql.FindSavedFilter(ctx, client, id)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", id).Msg("FindSavedFilter: Skipping")
			continue
		}

		savedFilter := savedFilterResponse.FindSavedFilter.SavedFilterParts

		if savedFilter.Mode != gql.FilterModeScenes {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("mode", string(savedFilter.Mode)).Msg("Unsupported filter mode, skipping")
			continue
		}

		filter, err := stash.ParseJsonEncodedFilter(savedFilter.Filter)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", id).RawJSON("jsonFilter", []byte(savedFilter.Filter)).Msg("ParseJsonEncodedFilter: Skipping")
			continue
		}

		scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filter.SceneFilter, filter.SortBy, filter.SortDir)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).
				Str("filterId", savedFilter.Id).
				RawJSON("jsonFilter", []byte(savedFilter.Filter)).
				RawJSON("parsedFilter", []byte(util.AsJsonStr(filter))).
				Msg("FindScenesByFilter: Skipping")
			continue
		}

		section := Section{
			Name:     getFilterName(prefix, savedFilter),
			FilterId: savedFilter.Id,
		}
		for _, s := range scenesResponse.FindScenes.Scenes {
			section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
		}
		*destination = append(*destination, section)

		log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Int("videoCount", len(section.PreviewPartsList)).Msg("Section added from Front Page")
	}

	return nil
}

func sectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string, destination *[]Section) error {
	savedFiltersResponse, err := gql.FindSavedFilters(ctx, client)
	if err != nil {
		return fmt.Errorf("FindSavedFilters: %w", err)
	}

	for _, savedFilter := range savedFiltersResponse.FindSavedFilters {
		if savedFilter.Mode != gql.FilterModeScenes {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Str("mode", string(savedFilter.Mode)).Msg("Unsupported filter mode, skipping")
			continue
		}

		if containsSavedFilterId(savedFilter.Id, *destination) {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("Filter has already been added, skipping")
			continue
		}

		filter, err := stash.ParseJsonEncodedFilter(savedFilter.Filter)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", savedFilter.Id).RawJSON("jsonFilter", []byte(savedFilter.Filter)).Msg("ParseJsonEncodedFilter: Skipping")
			continue
		}

		scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filter.SceneFilter, filter.SortBy, filter.SortDir)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).
				Str("filterId", savedFilter.Id).
				RawJSON("jsonFilter", []byte(savedFilter.Filter)).
				RawJSON("parsedFilter", []byte(util.AsJsonStr(filter))).
				Msg("FindScenesByFilter: Skipping")
			continue
		}
		if len(scenesResponse.FindScenes.Scenes) == 0 {
			log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Msg("0 videos, skipping")
			continue
		}

		section := Section{
			Name:     getFilterName(prefix, savedFilter.SavedFilterParts),
			FilterId: savedFilter.Id,
		}

		for _, s := range scenesResponse.FindScenes.Scenes {
			section.PreviewPartsList = append(section.PreviewPartsList, s.ScenePreviewParts)
		}
		*destination = append(*destination, section)

		log.Ctx(ctx).Debug().Str("filterId", savedFilter.Id).Str("filterName", savedFilter.Name).Int("videoCount", len(section.PreviewPartsList)).Msg("Section added from Saved Filter")
	}

	return nil
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
