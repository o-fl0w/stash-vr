package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
)

func destroy(ctx context.Context, client graphql.Client, sceneId string) error {
	_, err := gql.SceneDestroy(ctx, client, sceneId)
	if err != nil {
		return fmt.Errorf("SceneDestroy: %w", err)
	}
	log.Ctx(ctx).Debug().Str("sceneId", sceneId).Msg("Requested stash to delete scene, file and generated content")
	return nil
}
