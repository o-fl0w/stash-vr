package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
)

func FindFrontPageSavedFilterIds(ctx context.Context, client graphql.Client) ([]string, error) {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	var savedFilters []string
	frontPageFilters := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	for _, _filter := range frontPageFilters {
		filter := _filter.(map[string]interface{})
		typeName := filter["__typename"].(string)
		if typeName == "SavedFilter" {
			savedFilters = append(savedFilters, fmt.Sprintf("%.f", filter["savedFilterId"].(float64)))
		} else {
			log.Warn().Str("type", typeName).Msg("Unsupported filter type in front page, skipping")
		}
	}

	return savedFilters, nil
}
