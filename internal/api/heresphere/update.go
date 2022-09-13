package heresphere

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"math"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"strings"
)

type UpdateVideoData struct {
	Rating     *float32 `json:"rating,omitempty"`
	IsFavorite *bool    `json:"isFavorite,omitempty"`
	Tags       *[]Tag   `json:"tags,omitempty"`
	DeleteFile *bool    `json:"deleteFile,omitempty"`
}

func (v UpdateVideoData) IsUpdateRequest() bool {
	return v.Rating != nil || v.IsFavorite != nil || v.Tags != nil
}

func (v UpdateVideoData) IsDeleteRequest() bool {
	return v.DeleteFile != nil && *v.DeleteFile
}

type metadata struct {
	tagIds       *[]string
	studioId     *string
	performerIds *[]string
	markers      *[]sceneMarker
}

type sceneMarker struct {
	tag   string
	title string
	start float64
}

func update(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	log.Ctx(ctx).Trace().Str("sceneId", sceneId).Interface("update req", updateReq).Send()

	updateRating(ctx, client, sceneId, updateReq)
	updateFavorite(ctx, client, sceneId, updateReq)
	updateMetadata(ctx, client, sceneId, updateReq)
}

func updateRating(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	if updateReq.Rating == nil {
		return
	}
	var newRating int
	if *updateReq.Rating == 0.5 {
		//special case to set zero rating
		newRating = 0
	} else {
		newRating = int(math.Ceil(float64(*updateReq.Rating)))
	}

	_, err := gql.SceneUpdateRating(ctx, client, sceneId, newRating)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Int("new rating", newRating).Msg("Failed to update rating")
		return
	}
	log.Ctx(ctx).Debug().Str("sceneId", sceneId).Int("new rating", newRating).Msg("Rating updated")
}

func updateFavorite(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	if updateReq.IsFavorite == nil {
		return
	}

	isFavoriteRequested := *updateReq.IsFavorite

	favoriteTagName := config.Get().FavoriteTag
	favoriteTagId, err := stash.FindOrCreateTag(ctx, client, favoriteTagName)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Str("tag name", favoriteTagName).Msg("Failed to update favorite: FindOrCreateTag")
		return
	}

	response, err := gql.FindSceneTags(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Msg("Failed to update favorite: FindSceneTags")
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
		log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Interface("tags", newTagIds).Msg("Failed to update favorite: SceneUpdateTags")
	}
}

func updateMetadata(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	input := metadataFromUpdateRequest(ctx, client, updateReq)
	if input.tagIds != nil {
		_, err := gql.SceneUpdateTags(ctx, client, sceneId, *input.tagIds)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateTags")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Interface("new tags", *input.tagIds).Msg("Tags updated")
	}
	if input.studioId != nil {
		_, err := gql.SceneUpdateStudio(ctx, client, sceneId, *input.studioId)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateStudio")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Interface("new studio", *input.studioId).Msg("Studio updated")
	}
	if input.performerIds != nil {
		_, err := gql.SceneUpdatePerformers(ctx, client, sceneId, *input.performerIds)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdatePerformers")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Interface("new performers", *input.performerIds).Msg("Performers updated")
	}
	if input.markers != nil {
		setSceneMarkers(ctx, client, sceneId, *input.markers)
	}
}

func metadataFromUpdateRequest(ctx context.Context, client graphql.Client, updateReq UpdateVideoData) metadata {
	input := metadata{}

	if updateReq.Tags != nil {
		for _, tagReq := range *updateReq.Tags {
			tagType, tagName, isCategorized := strings.Cut(tagReq.Name, ":")

			if isCategorized && legendTag.IsMatch(tagType) {
				id, err := stash.FindOrCreateTag(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("FindOrCreateTag")
				}
				if input.tagIds == nil {
					input.tagIds = &[]string{}
				}
				*input.tagIds = append(*input.tagIds, id)
			} else if isCategorized && legendStudio.IsMatch(tagType) {
				id, err := stash.FindOrCreateStudio(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("FindOrCreateStudio")
				}
				input.studioId = &id
			} else if isCategorized && legendPerformer.IsMatch(tagType) {
				id, err := stash.FindOrCreatePerformer(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("FindOrCreatePerformer")
				}
				if input.performerIds == nil {
					input.performerIds = &[]string{}
				}
				*input.performerIds = append(*input.performerIds, id)
			} else {
				var markerTitle string
				markerPrimaryTag := tagType
				if isCategorized {
					markerTitle = tagName
				}
				if input.markers == nil {
					input.markers = &[]sceneMarker{}
				}
				*input.markers = append(*input.markers, sceneMarker{
					tag:   markerPrimaryTag,
					title: markerTitle,
					start: float64(tagReq.Start) / 1000,
				})
			}
		}
	}
	return input
}

func setSceneMarkers(ctx context.Context, client graphql.Client, sceneId string, markers []sceneMarker) {
	response, err := gql.FindSceneMarkers(ctx, client, sceneId)
	if err != nil {
		log.Warn().Err(err).Str("sceneId", sceneId).Msg("FindSceneMarkers")
		return
	}
	for _, smt := range response.SceneMarkerTags {
		for _, sm := range smt.Scene_markers {
			if _, err := gql.SceneMarkerDestroy(ctx, client, sm.Id); err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Str("sceneMarkerId", sm.Id).Str("sceneMarkerTitle", sm.Title).Msg("Failed to delete scene marker")
				continue
			}
		}
	}
	for _, m := range markers {
		tagId, err := stash.FindOrCreateTag(ctx, client, m.tag)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Msg("FindOrCreateTag")
			continue
		}
		_, err = gql.SceneMarkerCreate(ctx, client, sceneId, tagId, m.start, m.title)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Interface("marker", m).Msg("SceneMarkerCreate")
			continue
		}
	}
}
