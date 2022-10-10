package internal

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const sFilterList = "Filter List"

func SectionsByFilterIds(ctx context.Context, client graphql.Client, prefix string, filterIds []string) ([]section.Section, error) {
	savedFilters := stash.FindFiltersById(ctx, client, filterIds)

	sections := util.Transform[gql.SavedFilterParts, section.Section](func(savedFilter gql.SavedFilterParts) *section.Section {
		s, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			filterLogger(ctx, savedFilter, sFilterList).Warn().Err(err).Msg("Filter skipped")
			return nil
		}
		if len(s.PreviewPartsList) == 0 {
			filterLogger(ctx, savedFilter, sFilterList).Debug().Msg("Filter skipped: 0 scenes")
			return nil
		}
		sectionLogger(ctx, savedFilter, sFilterList, s).Debug().Msg("Section built")
		return &s
	}).Ordered(savedFilters)

	return sections, nil
}
