package internal

import (
	"context"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
)

func sourceLogContext(ctx context.Context, source string) context.Context {
	return log.Ctx(ctx).With().
		Str("source", source).Logger().WithContext(ctx)
}

func filterLogContext(ctx context.Context, filter gql.SavedFilterParts) context.Context {
	return log.Ctx(ctx).With().
		Str("filterId", filter.Id).Str("filterName", filter.Name).Str("filterMode", string(filter.Mode)).
		Logger().WithContext(ctx)
}

func sectionLogContext(ctx context.Context, section section.Section) context.Context {
	return log.Ctx(ctx).With().
		Str("section", section.Name).Int("scenes", len(section.Scene)).
		Logger().WithContext(ctx)
}
