package heresphere

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
)

func destroy(ctx context.Context, client graphql.Client, sceneId string) {
	if _, err := gql.SceneDestroy(ctx, client, sceneId); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to destroy scene")
		return
	}
	log.Ctx(ctx).Debug().Msg("Destroy scene request sent to Stash")
}
