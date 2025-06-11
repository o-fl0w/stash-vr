package stash

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strconv"
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

	switch minPlayPercent.(type) {
	case float64:
		return minPlayPercent.(float64)
	case string:
		v, err := strconv.Atoi(minPlayPercent.(string))
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Interface("config.minimumPlayPercent", minPlayPercent).Msg("Failed to parse Stash config.minimumPlayPercent")
			return 0
		}
		return float64(v)
	default:
		log.Ctx(ctx).Warn().Interface("config.minimumPlayPercent", minPlayPercent).Msg("Failed to parse Stash config.minimumPlayPercent: Unsupported format")
		return 0
	}
}
