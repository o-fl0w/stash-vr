package heresphere

import (
	"context"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
)

type updateVideoData struct {
	Rating     *float32 `json:"rating,omitempty"`
	IsFavorite *bool    `json:"isFavorite,omitempty"`
	Tags       *[]tag   `json:"tags,omitempty"`
	DeleteFile *bool    `json:"deleteFile,omitempty"`
}

func (v updateVideoData) isUpdateRequest() bool {
	return v.Rating != nil || v.IsFavorite != nil || v.Tags != nil
}

func (v updateVideoData) isDeleteRequest() bool {
	return v.DeleteFile != nil && *v.DeleteFile
}

func update(ctx context.Context, client graphql.Client, sceneId string, updateReq updateVideoData) {
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

func updateRating(ctx context.Context, client graphql.Client, sceneId string, rating float32) {
	var newRating int = int(rating*20 + 0.5)
	_, err := gql.SceneUpdateRating(ctx, client, sceneId, newRating)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Int("rating", newRating).Msg("Failed to update rating")
		return
	}

	log.Ctx(ctx).Debug().Int("rating", newRating).Msg("Updated rating")
}

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

	newTagIds := make([]string, 0, len(response.FindScene.Tags)+1)
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

func setMarkers(ctx context.Context, client graphql.Client, sceneId string, markers []marker) {
	if !config.Get().IsSyncMarkersAllowed {
		log.Ctx(ctx).Info().Msg("Sync markers requested but is disabled in config, ignoring request")
		return
	}
	response, err := gql.FindSceneMarkers(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("setMarkers: FindSceneMarkers")
		return
	}
	for _, smt := range response.SceneMarkerTags {
		for _, sm := range smt.Scene_markers {
			if _, err := gql.SceneMarkerDestroy(ctx, client, sm.Id); err != nil {
				log.Ctx(ctx).Warn().Err(err).
					Str("id", sm.Id).Str("title", sm.Title).Str("tag", sm.Primary_tag.Name).Msg("setMarkers: SceneMarkerDestroy")
				continue
			}
			log.Ctx(ctx).Debug().Str("id", sm.Id).Str("title", sm.Title).Str("tag", sm.Primary_tag.Name).Msg("Marker deleted")
		}
	}
	for _, m := range markers {
		tagId, err := stash.FindOrCreateTag(ctx, client, m.tag)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("title", m.title).Str("tag", m.tag).Msg("setMarkers: FindOrCreateTag")
			continue
		}
		createResponse, err := gql.SceneMarkerCreate(ctx, client, sceneId, tagId, m.start, m.title)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("tagId", tagId).Interface("marker", m).Msg("setMarkers: SceneMarkerCreate")
			continue
		}
		log.Ctx(ctx).Debug().Str("id", createResponse.SceneMarkerCreate.Id).Str("title", m.title).Str("tag", m.tag).Msg("Marker created")
	}
}

type requestDetails struct {
	tagIds          []string
	studioId        string
	performerIds    []string
	markers         []marker
	incrementO      bool
	toggleOrganized bool
}

type marker struct {
	tag   string
	title string
	start float64
}

func parseUpdateRequestTags(ctx context.Context, client graphql.Client, tags []tag) requestDetails {
	request := requestDetails{}

	for _, tagReq := range tags {
		if strings.HasPrefix(tagReq.Name, "!") {
			cmd := tagReq.Name[1:]
			if internal.LegendOCount.IsMatch(cmd) {
				request.incrementO = true
				continue
			} else if internal.LegendOrganized.IsMatch(cmd) {
				request.toggleOrganized = true
				continue
			}
		}

		if tagReq.Name == "" {
			continue
		}

		tagType, tagName, isCategorized := strings.Cut(tagReq.Name, ":")

		switch {
		case isCategorized && internal.LegendTag.IsMatch(tagType):
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty tag name, skipping")
				continue
			}
			id, err := stash.FindOrCreateTag(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create tag")
				continue
			}
			request.tagIds = append(request.tagIds, id)
		case isCategorized && internal.LegendStudio.IsMatch(tagType):
			if tagName == "" {
				continue
			}
			id, err := stash.FindOrCreateStudio(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create studio")
				continue
			}
			request.studioId = id
		case isCategorized && internal.LegendPerformer.IsMatch(tagType):
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty performer name, skipping")
				continue
			}
			id, err := stash.FindOrCreatePerformer(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create performer")
				continue
			}
			request.performerIds = append(request.performerIds, id)
		case isCategorized && (internal.LegendMovie.IsMatch(tagType) || internal.LegendOCount.IsMatch(tagType) || internal.LegendOrganized.IsMatch(tagType)):
			log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Tag type is reserved, skipping")
			continue
		default:
			if tagType == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty marker primary tag, skipping")
				continue
			}
			request.markers = append(request.markers, marker{
				tag:   tagType,
				title: tagName,
				start: tagReq.Start / 1000,
			})
		}
	}

	return request
}
