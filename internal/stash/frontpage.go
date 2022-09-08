package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
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
		}
	}

	return savedFilters, nil
}
