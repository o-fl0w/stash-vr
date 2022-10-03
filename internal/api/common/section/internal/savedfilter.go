package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/stash/filter/scenefilter"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
)

func SectionFromSavedFilter(ctx context.Context, client graphql.Client, prefix string, savedFilter gql.SavedFilterParts) (section.Section, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return section.Section{}, fmt.Errorf("unsupported filter mode")
	}

	filterQuery, err := scenefilter.ParseJsonEncodedFilter(savedFilter.Filter)
	if err != nil {
		return section.Section{}, fmt.Errorf("ParseJsonEncodedFilter: %w", err)
	}

	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filterQuery.SceneFilter, &filterQuery.FilterOpts)
	if err != nil {
		return section.Section{}, fmt.Errorf("FindScenesByFilter savedFilter=%+v parsedFilter=%+v: %w", savedFilter.Filter, util.AsJsonStr(filterQuery), err)
	}

	s := section.Section{
		Name:             getSectionName(prefix, savedFilter),
		FilterId:         savedFilter.Id,
		PreviewPartsList: make([]gql.ScenePreviewParts, len(scenesResponse.FindScenes.Scenes)),
	}

	for i, v := range scenesResponse.FindScenes.Scenes {
		s.PreviewPartsList[i] = v.ScenePreviewParts
	}
	return s, nil
}

func FindFiltersById(ctx context.Context, client graphql.Client, filterIds []string) []gql.SavedFilterParts {
	var filters []gql.SavedFilterParts

	for _, filterId := range filterIds {
		savedFilterResponse, err := gql.FindSavedFilter(ctx, client, filterId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("FindFiltersById: FindSavedFilter: %w", err)).Str("filterId", filterId).Msg("Skipped filter")
			continue
		}
		if savedFilterResponse.FindSavedFilter == nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("FindFiltersById: FindSavedFilter: Filter not found")).Str("filterId", filterId).Msg("Skipped filter")
			continue
		}
		filters = append(filters, savedFilterResponse.FindSavedFilter.SavedFilterParts)
	}

	return filters
}

func getSectionName(prefix string, fp gql.SavedFilterParts) string {
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
