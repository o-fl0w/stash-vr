package savedfilters

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	internal2 "stash-vr/internal/section/internal"
	"stash-vr/internal/section/model"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const sourceSavedFilters = "Saved Filters"

func Sections(ctx context.Context, client graphql.Client, prefix string) ([]model.Section, error) {

	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	savedFilters := make([]gql.SavedFilterParts, len(savedFiltersResponse.FindSavedFilters))
	for i, s := range savedFiltersResponse.FindSavedFilters {
		savedFilters[i] = s.SavedFilterParts
	}

	sections := util.Transform[gql.SavedFilterParts, model.Section](func(savedFilter gql.SavedFilterParts) *model.Section {
		s, err := internal2.SectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			internal2.FilterLogger(ctx, savedFilter, sourceSavedFilters).Warn().Err(err).Msg("Filter skipped")
			return nil
		}
		if len(s.PreviewPartsList) == 0 {
			internal2.FilterLogger(ctx, savedFilter, sourceSavedFilters).Debug().Msg("Filter skipped: 0 scenes")
			return nil
		}
		internal2.SectionLogger(ctx, savedFilter, sourceSavedFilters, s).Debug().Msg("Section built")
		return &s
	}).Ordered(savedFilters)

	return sections, nil
}
