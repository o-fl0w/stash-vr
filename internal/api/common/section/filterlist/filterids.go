package filterlist

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/api/common/section/internal"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const source = "Filter List"

func Sections(ctx context.Context, client graphql.Client, prefix string, filterIds []string) ([]section.Section, error) {

	savedFilters := internal.FindFiltersById(ctx, client, filterIds)

	sections := util.Transform[gql.SavedFilterParts, section.Section](func(savedFilter gql.SavedFilterParts) *section.Section {
		s, err := internal.SectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			internal.FilterLogger(ctx, savedFilter, source).Warn().Err(err).Msg("Filter skipped")
			return nil
		}
		if len(s.PreviewPartsList) == 0 {
			internal.FilterLogger(ctx, savedFilter, source).Debug().Msg("Filter skipped: 0 scenes")
			return nil
		}
		internal.SectionLogger(ctx, savedFilter, source, s).Debug().Msg("Section built")
		return &s
	}).Ordered(savedFilters)

	return sections, nil
}
