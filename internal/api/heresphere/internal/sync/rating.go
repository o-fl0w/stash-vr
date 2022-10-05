package sync

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"math"
	"stash-vr/internal/stash/gql"
)

func updateRating(ctx context.Context, client graphql.Client, sceneId string, rating float32) {
	var newRating int
	if rating == 0.5 {
		//special case to set zero rating
		newRating = 0
	} else {
		newRating = int(math.Ceil(float64(rating)))
	}

	_, err := gql.SceneUpdateRating(ctx, client, sceneId, newRating)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Int("rating", newRating).Msg("Failed to update rating")
		return
	}

	log.Ctx(ctx).Debug().Int("rating", newRating).Msg("Updated rating")
}
