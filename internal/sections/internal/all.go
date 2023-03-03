package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/logger"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/filter"
	"stash-vr/internal/stash/gql"
)

func SectionWithAllScenes(ctx context.Context, client graphql.Client) (section.Section, error) {
	fq := filter.Filter{
		FilterOpts: gql.FindFilterType{
			Per_page:  -1,
			Direction: gql.SortDirectionEnumAsc,
		},
		SceneFilter: gql.SceneFilterType{},
	}

	scenesResponse, err := gql.FindScenePreviewsByFilter(ctx, client, &fq.SceneFilter, &fq.FilterOpts)
	if err != nil {
		return section.Section{}, fmt.Errorf("FindScenePreviewsByFilter filter=%+v: %w", logger.AsJsonStr(fq), err)
	}

	s := section.Section{
		Name:  "All",
		Scene: make([]gql.ScenePreviewParts, len(scenesResponse.FindScenes.Scenes)),
	}

	for i, v := range scenesResponse.FindScenes.Scenes {
		s.Scene[i] = v.ScenePreviewParts
	}
	return s, nil
}
