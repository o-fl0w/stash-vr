package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash"
)

func SectionsByFrontpage(ctx context.Context, client graphql.Client, prefix string) ([]section.Section, error) {
	filterIds, err := stash.FindSavedFilterIdsByFrontPage(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedFilterIdsByFrontPage: %w", err)
	}

	savedFilters := stash.FindFiltersById(ctx, client, filterIds)

	sections := sectionFromSavedFilterFuncBuilder(ctx, client, prefix, "Front Page").Ordered(savedFilters)

	return sections, nil
}
