package internal

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash"
)

func SectionsByFilterIds(ctx context.Context, client graphql.Client, prefix string, filterIds []string) ([]section.Section, error) {
	savedFilters := stash.FindFiltersById(ctx, client, filterIds)

	sections := sectionFromSavedFilterFuncBuilder(ctx, client, prefix, "Filter List").Ordered(savedFilters)

	return sections, nil
}
