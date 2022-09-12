package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strconv"
)

func FindSavedSceneFiltersByFrontPage(ctx context.Context, client graphql.Client) ([]gql.SavedFilterParts, error) {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	var savedSceneFilters []gql.SavedFilterParts
	frontPageFilters := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	for _, _filter := range frontPageFilters {
		filter := _filter.(map[string]interface{})
		typeName := filter["__typename"].(string)
		if typeName != "SavedFilter" {
			log.Ctx(ctx).Info().Err(fmt.Errorf("unsupported filter type '%s'", typeName)).Msg("Skipped filter: FindSavedSceneFiltersByFrontPage")
			continue
		}
		filterId := strconv.Itoa(int(filter["savedFilterId"].(float64)))

		savedFilterResponse, err := gql.FindSavedFilter(ctx, client, filterId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("filterId", filterId).Msg("Skipped filter: FindSavedSceneFiltersByFrontPage: FindSavedFilter")
			continue
		}
		if savedFilterResponse.FindSavedFilter.Mode != gql.FilterModeScenes {
			log.Ctx(ctx).Debug().Str("filterId", filterId).
				Str("mode", string(savedFilterResponse.FindSavedFilter.Mode)).
				Msg("FindSavedSceneFiltersByFrontPage: FindSavedFilter: Not a scene filter, skipped")
			continue
		}
		savedSceneFilters = append(savedSceneFilters, savedFilterResponse.FindSavedFilter.SavedFilterParts)
	}

	return savedSceneFilters, nil
}
