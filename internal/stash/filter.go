package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strconv"
)

func FindFiltersById(ctx context.Context, client graphql.Client, filterIds []string) []gql.SavedFilterParts {
	filters := make([]gql.SavedFilterParts, 0, len(filterIds))

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

func FindSavedFilterIdsByFrontPage(ctx context.Context, client graphql.Client) ([]string, error) {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	frontPageContent := configurationResponse.Configuration.Ui["frontPageContent"]
	if frontPageContent == nil {
		log.Ctx(ctx).Info().Msg("No frontpage content found")
		return nil, nil
	}

	frontPageFilters := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	filterIds := make([]string, 0, len(frontPageFilters))
	for _, _filter := range frontPageFilters {
		filter := _filter.(map[string]interface{})
		typeName := filter["__typename"].(string)
		if typeName != "SavedFilter" {
			log.Ctx(ctx).Debug().Str("type", typeName).Msg("Filter skipped: Unsupported filter type on front page: Only user created saved scene filters are supported.")
			continue
		}
		fid := filter["savedFilterId"]
		if fid == nil {
			continue
		}
		filterId, ok := fid.(string)
		if !ok {
			filterId = strconv.Itoa(int(fid.(float64)))
		}
		filterIds = append(filterIds, filterId)
	}

	return filterIds, nil
}
