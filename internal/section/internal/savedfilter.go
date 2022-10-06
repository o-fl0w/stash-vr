package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/section"
	"stash-vr/internal/stash/filter"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
)

func sectionFromSavedFilter(ctx context.Context, client graphql.Client, prefix string, savedFilter gql.SavedFilterParts) (section.Section, error) {
	filterQuery, err := filter.SavedFilterToSceneFilter(savedFilter)
	if err != nil {
		return section.Section{}, fmt.Errorf("SavedFilterToSceneFilter: %w", err)
	}

	scenesResponse, err := gql.FindScenePreviewsByFilter(ctx, client, &filterQuery.SceneFilter, &filterQuery.FilterOpts)
	if err != nil {
		return section.Section{}, fmt.Errorf("FindScenePreviewsByFilter savedFilter=%+v parsedFilter=%+v: %w", savedFilter.Filter, util.AsJsonStr(filterQuery), err)
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
