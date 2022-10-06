package internal

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/section/model"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

func FilterLogger(ctx context.Context, filter gql.SavedFilterParts, source string) *zerolog.Logger {
	return util.Ptr(log.Ctx(ctx).With().
		Str("filterId", filter.Id).Str("filterName", filter.Name).Interface("filterMode", filter.Mode).
		Str("source", source).Logger())
}

func SectionLogger(ctx context.Context, filter gql.SavedFilterParts, source string, section model.Section) *zerolog.Logger {
	return util.Ptr(log.Ctx(ctx).With().
		Str("filterId", filter.Id).Str("filterName", filter.Name).Interface("filterMode", filter.Mode).
		Str("source", source).
		Str("section", section.Name).Int("scenes", len(section.PreviewPartsList)).Logger())
}
