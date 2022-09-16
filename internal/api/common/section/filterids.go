package section

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const sourceFilterIds = "Filter Ids"

func SectionsByFilterIds(ctx context.Context, client graphql.Client, prefix string, filterIds []string) ([]Section, error) {
	savedFilters := findFiltersById(ctx, client, filterIds)

	sections := util.Transform[gql.SavedFilterParts, Section](func(savedFilter gql.SavedFilterParts) (Section, error) {
		section, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			filterLogger(ctx, savedFilter, sourceFilterIds).Warn().Err(err).Msg("Skipped filter: sectionsByFilterIds: sectionFromSavedFilter")
			return Section{}, err
		}
		sectionLogger(ctx, savedFilter, sourceFilterIds, section).Debug().Msg("Section built")
		return section, nil
	}).Ordered(savedFilters)

	return sections, nil
}

func findFiltersById(ctx context.Context, client graphql.Client, filterIds []string) []gql.SavedFilterParts {
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
