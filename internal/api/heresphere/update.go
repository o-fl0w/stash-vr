package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"math"
	"stash-vr/internal/api/common"
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
	tagIds          []string
	studioId        string
	performerIds    []string
	markers         []sceneMarker
	incrementO      bool
	toggleOrganized bool
}

type sceneMarker struct {
	tag   string
	title string
	start float64
}

func update(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	log.Ctx(ctx).Debug().Interface("data", updateReq).Msg("Update request")

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
		log.Ctx(ctx).Warn().Err(fmt.Errorf("updateRating: SceneUpdateRating: %w", err)).Int("new rating", newRating).Send()
		return
	}
	log.Ctx(ctx).Trace().Int("rating", newRating).Msg("Scene: Rating updated")
}

func updateFavorite(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	if updateReq.IsFavorite == nil {
		return
	}

	isFavoriteRequested := *updateReq.IsFavorite

	favoriteTagName := config.Get().FavoriteTag
	favoriteTagId, err := stash.FindOrCreateTag(ctx, client, favoriteTagName)
	if err != nil {
		log.Ctx(ctx).Warn().Err(fmt.Errorf("updateFavorite: FindOrCreateTag: %w", err)).Str("tag name", favoriteTagName).Send()
		return
	}

	response, err := gql.FindSceneTags(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(fmt.Errorf("updateFavorite: FindSceneTags: %w", err)).Send()
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
		log.Ctx(ctx).Warn().Err(fmt.Errorf("updateFavorite: SceneUpdateTags: %w", err)).Interface("tags", newTagIds).Send()
	}
	log.Ctx(ctx).Trace().Bool("isFavorite", isFavoriteRequested).Msg("Scene: Favorite updated")
}

func updateMetadata(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	if updateReq.Tags == nil {
		return
	}

	input := metadataFromUpdateRequestTags(ctx, client, *updateReq.Tags)

	var err error

	_, err = gql.SceneUpdateTags(ctx, client, sceneId, input.tagIds)
	if err != nil {
		log.Ctx(ctx).Warn().Err(fmt.Errorf("updateMetadata: SceneUpdateTags: %w", err)).Interface("input", input.tagIds).Send()
	} else {
		log.Ctx(ctx).Debug().Interface("tags", input.tagIds).Msg("Scene: Tags updated")
	}

	if input.studioId == "" {
		_, err = gql.SceneClearStudio(ctx, client, sceneId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("updateMetadata: SceneClearStudio: %w", err)).Send()
		} else {
			log.Ctx(ctx).Debug().Interface("studio", input.studioId).Msg("Scene: Studio cleared")
		}
	} else {
		_, err = gql.SceneUpdateStudio(ctx, client, sceneId, input.studioId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("updateMetadata: SceneUpdateStudio: %w", err)).Interface("input", input.studioId).Send()
		} else {
			log.Ctx(ctx).Debug().Interface("studio", input.studioId).Msg("Scene: Studio updated")
		}
	}

	_, err = gql.SceneUpdatePerformers(ctx, client, sceneId, input.performerIds)
	if err != nil {
		log.Ctx(ctx).Warn().Err(fmt.Errorf("updateMetadata: SceneUpdatePerformers: %w", err)).Interface("input", input.performerIds).Send()
	} else {
		log.Ctx(ctx).Debug().Interface("performers", input.performerIds).Msg("Scene: Performers updated")
	}

	setSceneMarkers(ctx, client, sceneId, input.markers)

	if input.incrementO {
		newCount, err := gql.SceneIncrementO(ctx, client, sceneId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("updateMetadata: SceneIncrementO: %w", err)).Interface("increment", input.incrementO).Send()
		} else {
			log.Ctx(ctx).Debug().Interface("O-counter", newCount).Msg("Scene: O-counter updated")
		}
	}
	if input.toggleOrganized {
		newOrganized, err := stash.SceneToggleOrganized(ctx, client, sceneId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("updateMetadata: SceneToggleOrganized: %w", err)).Interface("toggle", input.toggleOrganized).Send()
		} else {
			log.Ctx(ctx).Debug().Interface("Organized", newOrganized).Msg("Scene: Organized updated")
		}
	}
}

func metadataFromUpdateRequestTags(ctx context.Context, client graphql.Client, tags []Tag) metadata {
	input := metadata{}

	for _, tagReq := range tags {
		if strings.HasPrefix(tagReq.Name, "!") {
			cmd := tagReq.Name[1:]
			if common.LegendOCount.IsMatch(cmd) {
				input.incrementO = true
				continue
			} else if common.LegendOrganized.IsMatch(cmd) {
				input.toggleOrganized = true
				continue
			}
		}

		tagType, tagName, isCategorized := strings.Cut(tagReq.Name, ":")

		if isCategorized && common.LegendTag.IsMatch(tagType) {
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty tag name, skipping")
				continue
			}
			id, err := stash.FindOrCreateTag(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(fmt.Errorf("metadataFromUpdateRequest: FindOrCreateTag: %w", err)).Str("request", tagReq.Name).Send()
				continue
			}
			input.tagIds = append(input.tagIds, id)
		} else if isCategorized && common.LegendStudio.IsMatch(tagType) {
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty studio name, skipping")
				continue
			}
			id, err := stash.FindOrCreateStudio(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(fmt.Errorf("metadataFromUpdateRequest: FindOrCreateStudio: %w", err)).Str("request", tagReq.Name).Send()
				continue
			}
			input.studioId = id
		} else if isCategorized && common.LegendPerformer.IsMatch(tagType) {
			if tagName == "" {
				log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Empty performer name, skipping")
				continue
			}
			id, err := stash.FindOrCreatePerformer(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(fmt.Errorf("metadataFromUpdateRequest: FindOrCreatePerformer: %w", err)).Str("request", tagReq.Name).Send()
				continue
			}
			input.performerIds = append(input.performerIds, id)
		} else if isCategorized && (common.LegendMovie.IsMatch(tagType) || common.LegendOCount.IsMatch(tagType) || common.LegendOrganized.IsMatch(tagType)) {
			log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Tag type is reserved, skipping")
			continue
		} else {
			var markerTitle string
			markerPrimaryTag := tagType
			if isCategorized {
				markerTitle = tagName
			}
			input.markers = append(input.markers, sceneMarker{
				tag:   markerPrimaryTag,
				title: markerTitle,
				start: float64(tagReq.Start) / 1000,
			})
		}
	}

	return input
}

func setSceneMarkers(ctx context.Context, client graphql.Client, sceneId string, markers []sceneMarker) {
	if !config.Get().HeresphereSyncMarkers {
		log.Ctx(ctx).Info().Bool(config.EnvKeyHeresphereSyncMarkers, config.Get().HeresphereSyncMarkers).Msg("Markers received from HereSphere but sync for markers is disabled, ignoring.")
		return
	}
	response, err := gql.FindSceneMarkers(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(fmt.Errorf("setSceneMarkers: FindSceneMarkers: %w", err)).Send()
		return
	}
	for _, smt := range response.SceneMarkerTags {
		for _, sm := range smt.Scene_markers {
			if _, err := gql.SceneMarkerDestroy(ctx, client, sm.Id); err != nil {
				log.Ctx(ctx).Warn().Err(fmt.Errorf("setSceneMarkers: SceneMarkerDestroy: %w", err)).
					Str("id", sm.Id).Str("title", sm.Title).Str("tag", sm.Primary_tag.Name).Msg("Failed to delete marker")
				continue
			}
			log.Ctx(ctx).Trace().Str("id", sm.Id).Str("title", sm.Title).Str("tag", sm.Primary_tag.Name).Msg("Marker deleted, will recreate...")
		}
	}
	for _, m := range markers {
		tagId, err := stash.FindOrCreateTag(ctx, client, m.tag)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("setSceneMarkers: FindOrCreateTag: %w", err)).Str("title", m.title).Str("tag", m.tag).Msg("Failed to create tag for marker")
			continue
		}
		createResponse, err := gql.SceneMarkerCreate(ctx, client, sceneId, tagId, m.start, m.title)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("setSceneMarkers: SceneMarkerCreate: %w", err)).Interface("marker", m).Msg("Failed to create marker")
			continue
		}
		log.Ctx(ctx).Trace().Str("id", createResponse.SceneMarkerCreate.Id).Str("title", m.title).Str("tag", m.tag).Msg("Marker created")
	}
}
