package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
)

func SectionsBySavedFilters(ctx context.Context, client graphql.Client, prefix string) ([]section.Section, error) {

	savedFiltersResponse, err := gql.FindSavedSceneFilters(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedSceneFilters: %w", err)
	}

	savedFilters := make([]gql.SavedFilterParts, len(savedFiltersResponse.FindSavedFilters))
	for i, s := range savedFiltersResponse.FindSavedFilters {
		savedFilters[i] = s.SavedFilterParts
	}

	sections := sectionFromSavedFilterFuncBuilder(ctx, client, prefix, "Saved Filters").Ordered(savedFilters)

	return sections, nil
}
