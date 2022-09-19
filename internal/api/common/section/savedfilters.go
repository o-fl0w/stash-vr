package section

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

const sourceSavedFilters = "Saved Filters"

func SectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string) ([]Section, error) {

	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	savedFilters := make([]gql.SavedFilterParts, len(savedFiltersResponse.FindSavedFilters))
	for i, s := range savedFiltersResponse.FindSavedFilters {
		savedFilters[i] = s.SavedFilterParts
	}

	sections := util.Transform[gql.SavedFilterParts, Section](func(savedFilter gql.SavedFilterParts) (Section, error) {
		section, err := sectionFromSavedFilter(ctx, client, prefix, savedFilter)
		if err != nil {
			filterLogger(ctx, savedFilter, sourceSavedFilters).Warn().Err(err).Msg("Skipped filter: sectionsBySavedFilters: sectionFromSavedFilter")
			return Section{}, err
		}
		sectionLogger(ctx, savedFilter, sourceSavedFilters, section).Debug().Msg("Section built")
		return section, nil
	}).Ordered(savedFilters)

	return sections, nil
}
