package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strconv"
)

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
			log.Ctx(ctx).Debug().Str("type", typeName).Msg("Filter skipped: Unsupported filter type on front page: Only user created SCENE filters are supported.")
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
