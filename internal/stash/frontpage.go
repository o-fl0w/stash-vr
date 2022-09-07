package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"net/http"
	"stash-vr/internal/stash/gql"
)

func FindFrontPageSavedFilterIds(ctx context.Context, serverUrl string) ([]string, error) {
	client := graphql.NewClient(serverUrl, http.DefaultClient)

	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	var savedFilters []string
	fpcs := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	for _, sfo := range fpcs {
		sfm := sfo.(map[string]interface{})
		savedFilters = append(savedFilters, fmt.Sprintf("%.f", sfm["savedFilterId"].(float64)))
	}

	return savedFilters, nil
}
