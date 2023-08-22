package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/logger"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/filter"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
)

type sectionFromSavedFilterFunc = util.Transform[gql.SavedFilterParts, section.Section]

var errNoScenesFound = errors.New("no scenes found")

func sectionFromSavedFilterFuncBuilder(ctx context.Context, client graphql.Client, prefix string, source string) sectionFromSavedFilterFunc {
	return func(savedFilter gql.SavedFilterParts) (section.Section, error) {
		ctx := sourceLogContext(filterLogContext(ctx, savedFilter), source)
		s, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Filter skipped")
			return section.Section{}, err
		}
		if len(s.Scenes) == 0 {
			log.Ctx(ctx).Debug().Msg("Filter skipped: 0 scenes")
			return section.Section{}, errNoScenesFound
		}
		ctx = sectionLogContext(ctx, s)
		log.Ctx(ctx).Debug().Msg("Section built")
		return s, nil
	}
}

func sectionFromSavedFilter(ctx context.Context, client graphql.Client, prefix string, savedFilter gql.SavedFilterParts) (section.Section, error) {
	filterQuery, err := filter.SavedFilterToSceneFilter(ctx, savedFilter)
	if err != nil {
		return section.Section{}, fmt.Errorf("SavedFilterToSceneFilter: %w", err)
	}

	scenesResponse, err := gql.FindScenePreviewsByFilter(ctx, client, &filterQuery.SceneFilter, &filterQuery.FilterOpts)
	if err != nil {
		return section.Section{}, fmt.Errorf("FindScenePreviewsByFilter savedFilter=%+v parsedFilter=%+v: %w", savedFilter.Filter, logger.AsJsonStr(filterQuery), err)
	}

	s := section.Section{
		Name:     getSectionName(prefix, savedFilter),
		FilterId: savedFilter.Id,
		Scenes:   make([]section.ScenePreview, len(scenesResponse.FindScenes.Scenes)),
	}

	for i, v := range scenesResponse.FindScenes.Scenes {
		s.Scenes[i] = section.ScenePreview{
			ScenePreviewParts: v.ScenePreviewParts,
		}
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
