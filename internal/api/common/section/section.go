package section

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func ContainsFilterId(id string, list []Section) bool {
	for _, v := range list {
		if id == v.FilterId {
			return true
		}
	}
	return false
}

func sectionFromSavedFilter(ctx context.Context, client graphql.Client, prefix string, savedFilter gql.SavedFilterParts) (Section, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return Section{}, fmt.Errorf("unsupported filter mode")
	}

	filterQuery, err := scenefilter.ParseJsonEncodedFilter(savedFilter.Filter)
	if err != nil {
		return Section{}, fmt.Errorf("ParseJsonEncodedFilter: %w", err)
	}

	scenesResponse, err := gql.FindScenesByFilter(ctx, client, &filterQuery.SceneFilter, &filterQuery.FilterOpts)
	if err != nil {
		return Section{}, fmt.Errorf("FindScenesByFilter savedFilter=%+v parsedFilter=%+v: %w", savedFilter.Filter, util.AsJsonStr(filterQuery), err)
	}

	if len(scenesResponse.FindScenes.Scenes) == 0 {
		return Section{}, fmt.Errorf("0 videos found")
	}

	section := Section{
		Name:             getSectionName(prefix, savedFilter),
		FilterId:         savedFilter.Id,
		PreviewPartsList: make([]gql.ScenePreviewParts, len(scenesResponse.FindScenes.Scenes)),
	}

	for i, s := range scenesResponse.FindScenes.Scenes {
		section.PreviewPartsList[i] = s.ScenePreviewParts
	}
	return section, nil
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

func filterLogger(ctx context.Context, filter gql.SavedFilterParts, source string) *zerolog.Logger {
	return util.Ptr(log.Ctx(ctx).With().
		Str("filterId", filter.Id).Str("filterName", filter.Name).Interface("filterMode", filter.Mode).
		Str("source", source).Logger())
}

func sectionLogger(ctx context.Context, filter gql.SavedFilterParts, source string, section Section) *zerolog.Logger {
	return util.Ptr(log.Ctx(ctx).With().
		Str("filterId", filter.Id).Str("filterName", filter.Name).Interface("filterMode", filter.Mode).
		Str("source", source).
		Str("section", section.Name).Int("links", len(section.PreviewPartsList)).Logger())
}
