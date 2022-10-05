package sync

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

func Update(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	log.Ctx(ctx).Debug().Interface("data", updateReq).Msg("Update request")

	if updateReq.Rating != nil {
		updateRating(ctx, client, sceneId, *updateReq.Rating)
	}

	if updateReq.IsFavorite != nil {
		updateFavorite(ctx, client, sceneId, *updateReq.IsFavorite)
	}

	if updateReq.Tags != nil {
		details := parseUpdateRequestTags(ctx, client, *updateReq.Tags)

		updateTags(ctx, client, sceneId, details.tagIds)
		updateStudio(ctx, client, sceneId, details.studioId)
		updatePerformers(ctx, client, sceneId, details.performerIds)

		if details.incrementO {
			incrementO(ctx, client, sceneId)
		}

		if details.toggleOrganized {
			toggleOrganized(ctx, client, sceneId)
		}

		setMarkers(ctx, client, sceneId, details.markers)
	}
}

func updateTags(ctx context.Context, client graphql.Client, sceneId string, tagIds []string) {
	if _, err := gql.SceneUpdateTags(ctx, client, sceneId, tagIds); err != nil {
		log.Ctx(ctx).Warn().Err(err).Interface("tagIds", tagIds).Msg("Failed to update tags")
		return
	}
	log.Ctx(ctx).Debug().Interface("tagIds", tagIds).Msg("Updated tags")
}

func updateStudio(ctx context.Context, client graphql.Client, sceneId string, studioId string) {
	if studioId == "" {
		if _, err := gql.SceneClearStudio(ctx, client, sceneId); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to clear studio")
			return
		}
	} else {
		if _, err := gql.SceneUpdateStudio(ctx, client, sceneId, studioId); err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("studioId", studioId).Msg("Failed to update studio")
			return
		}
	}
	log.Ctx(ctx).Debug().Str("studioId", studioId).Msg("Updated studio")
}

func updatePerformers(ctx context.Context, client graphql.Client, sceneId string, performerIds []string) {
	if _, err := gql.SceneUpdatePerformers(ctx, client, sceneId, performerIds); err != nil {
		log.Ctx(ctx).Warn().Err(err).Interface("performerIds", performerIds).Msg("Failed to update performers")
		return
	}
	log.Ctx(ctx).Debug().Interface("performerIds", performerIds).Msg("Updated performers")
}

func incrementO(ctx context.Context, client graphql.Client, sceneId string) {
	response, err := gql.SceneIncrementO(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to increment O-count")
		return
	}
	log.Ctx(ctx).Debug().Interface("O-count", response.SceneIncrementO).Msg("Incremented O-count")
}

func toggleOrganized(ctx context.Context, client graphql.Client, sceneId string) {
	newOrganized, err := stash.SceneToggleOrganized(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to toggle organized flag")
		return
	}
	log.Ctx(ctx).Debug().Bool("organized", newOrganized).Msg("Toggled organized flag")
}
