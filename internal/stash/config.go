package stash

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
)

func GetMinPlayPercent(ctx context.Context, client graphql.Client) float64 {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to retrieve stash configuration")
		return 0
	}

	minPlayPercent := configurationResponse.Configuration.Ui["minimumPlayPercent"]
	if minPlayPercent == nil {
		return 0
	}

	return minPlayPercent.(float64)
}
