package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const sSavedFilters = "Saved Filters"

func SectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string) ([]section.Section, error) {

	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	savedFilters := make([]gql.SavedFilterParts, len(savedFiltersResponse.FindSavedFilters))
	for i, s := range savedFiltersResponse.FindSavedFilters {
		savedFilters[i] = s.SavedFilterParts
	}

	sections := util.Transform[gql.SavedFilterParts, section.Section](func(savedFilter gql.SavedFilterParts) *section.Section {
		s, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			filterLogger(ctx, savedFilter, sSavedFilters).Warn().Err(err).Msg("Filter skipped")
			return nil
		}
		if len(s.PreviewPartsList) == 0 {
			filterLogger(ctx, savedFilter, sSavedFilters).Debug().Msg("Filter skipped: 0 scenes")
			return nil
		}
		sectionLogger(ctx, savedFilter, sSavedFilters, s).Debug().Msg("Section built")
		return &s
	}).Ordered(savedFilters)

	return sections, nil
}
