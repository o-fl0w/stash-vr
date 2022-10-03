package frontpage

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common/section"
	"stash-vr/internal/api/common/section/internal"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strconv"
)

const source = "Front Page"

func Sections(ctx context.Context, client graphql.Client, prefix string) ([]section.Section, error) {
	filterIds, err := findSavedFilterIdsByFrontPage(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("FindSavedFilterIdsByFrontPage: %w", err)
	}

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

func findSavedFilterIdsByFrontPage(ctx context.Context, client graphql.Client) ([]string, error) {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	var filterIds []string
	frontPageContent := configurationResponse.Configuration.Ui["frontPageContent"]
	if frontPageContent == nil {
		log.Ctx(ctx).Info().Msg("No frontpage content found")
		return nil, nil
	}
	frontPageFilters := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	for _, _filter := range frontPageFilters {
		filter := _filter.(map[string]interface{})
		typeName := filter["__typename"].(string)
		if typeName != "SavedFilter" {
			log.Ctx(ctx).Info().Str("type", typeName).Str("source", source).Msg("Filter skipped: Unsupported type. Only user created saved scene filters are supported")
			continue
		}

		filterId := strconv.Itoa(int(filter["savedFilterId"].(float64)))
		filterIds = append(filterIds, filterId)
	}

	return filterIds, nil
}
