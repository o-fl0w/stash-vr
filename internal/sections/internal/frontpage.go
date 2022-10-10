package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const sFrontPage = "Front Page"

func SectionsByFrontpage(ctx context.Context, client graphql.Client, prefix string) ([]section.Section, error) {
	filterIds, err := stash.FindSavedFilterIdsByFrontPage(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedFilterIdsByFrontPage: %w", err)
	}

	savedFilters := stash.FindFiltersById(ctx, client, filterIds)

	sections := util.Transform[gql.SavedFilterParts, section.Section](func(savedFilter gql.SavedFilterParts) *section.Section {
		s, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			filterLogger(ctx, savedFilter, sFrontPage).Warn().Err(err).Msg("Filter skipped")
			return nil
		}
		if len(s.PreviewPartsList) == 0 {
			filterLogger(ctx, savedFilter, sFrontPage).Debug().Msg("Filter skipped: 0 scenes")
			return nil
		}
		sectionLogger(ctx, savedFilter, sFrontPage, s).Debug().Msg("Section built")
		return &s
	}).Ordered(savedFilters)

	return sections, nil
}
