package sync

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

func updateFavorite(ctx context.Context, client graphql.Client, sceneId string, isFavoriteRequested bool) {
	favoriteTagName := config.Get().FavoriteTag

	if favoriteTagName == "" {
		log.Ctx(ctx).Info().Msg("Sync favorite requested but FAVORITE_TAG is empty, ignoring request")
		return
	}

	favoriteTagId, err := stash.FindOrCreateTag(ctx, client, favoriteTagName)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Bool("isFavorite", isFavoriteRequested).Str("tagName", favoriteTagName).Msg("Failed to update favorite: FindOrCreateTag")
		return
	}

	response, err := gql.FindSceneTags(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Bool("isFavorite", isFavoriteRequested).Msg("Failed to update favorite: FindSceneTags")
		return
	}

	var newTagIds []string
	var contains bool
	for _, t := range response.FindScene.Tags {
		if t.Id == favoriteTagId {
			contains = true
			if !isFavoriteRequested {
				continue
			}
		}
		newTagIds = append(newTagIds, t.Id)
	}
	if !contains && isFavoriteRequested {
		newTagIds = append(newTagIds, favoriteTagId)
	}

	if _, err := gql.SceneUpdateTags(ctx, client, sceneId, newTagIds); err != nil {
		log.Ctx(ctx).Warn().Err(err).Bool("isFavorite", isFavoriteRequested).Interface("newTagIds", newTagIds).Msg("Failed to update favorite: SceneUpdateTags")
		return
	}

	log.Ctx(ctx).Debug().Bool("isFavorite", isFavoriteRequested).Msg("Updated favorite")
}
