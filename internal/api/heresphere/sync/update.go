package sync

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/heresphere/proto"

	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"strings"
)

type UpdateVideoData struct {
	Rating     *float32     `json:"rating,omitempty"`
	IsFavorite *bool        `json:"isFavorite,omitempty"`
	Tags       *[]proto.Tag `json:"tags,omitempty"`
	DeleteFile *bool        `json:"deleteFile,omitempty"`
}

func (v UpdateVideoData) IsUpdateRequest() bool {
	return v.Rating != nil || v.IsFavorite != nil || v.Tags != nil
}

func (v UpdateVideoData) IsDeleteRequest() bool {
	return v.DeleteFile != nil && *v.DeleteFile
}

type requestDetails struct {
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

func Update(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	log.Ctx(ctx).Debug().Interface("data", updateReq).Msg("Update request")

	if updateReq.Rating != nil {
		updateRating(ctx, client, sceneId, *updateReq.Rating)
	}

	if updateReq.IsFavorite != nil {
		updateFavorite(ctx, client, sceneId, *updateReq.IsFavorite)
	}

	if updateReq.Tags != nil {
		request := parseUpdateRequestTags(ctx, client, *updateReq.Tags)

		updateTags(ctx, client, sceneId, request.tagIds)
		updateStudio(ctx, client, sceneId, request.studioId)
		updatePerformers(ctx, client, sceneId, request.performerIds)

		if request.incrementO {
			incrementO(ctx, client, sceneId)
		}

		if request.toggleOrganized {
			toggleOrganized(ctx, client, sceneId)
		}

		setMarkers(ctx, client, sceneId, request.markers)
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

func parseUpdateRequestTags(ctx context.Context, client graphql.Client, tags []proto.Tag) requestDetails {
	request := requestDetails{}

	for _, tagReq := range tags {
		if strings.HasPrefix(tagReq.Name, "!") {
			cmd := tagReq.Name[1:]
			if common.LegendOCount.IsMatch(cmd) {
				request.incrementO = true
				continue
			} else if common.LegendOrganized.IsMatch(cmd) {
				request.toggleOrganized = true
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
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create tag")
				continue
			}
			request.tagIds = append(request.tagIds, id)
		} else if isCategorized && common.LegendStudio.IsMatch(tagType) {
			if tagName == "" {
				continue
			}
			id, err := stash.FindOrCreateStudio(ctx, client, tagName)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("request", tagReq.Name).Msg("Failed to find or create studio")
				continue
			}
			request.studioId = id
		} else if isCategorized && common.LegendPerformer.IsMatch(tagType) {
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
		} else if isCategorized && (common.LegendMovie.IsMatch(tagType) || common.LegendOCount.IsMatch(tagType) || common.LegendOrganized.IsMatch(tagType)) {
			log.Ctx(ctx).Trace().Str("request", tagReq.Name).Msg("Tag type is reserved, skipping")
			continue
		} else {
			var markerTitle string
			markerPrimaryTag := tagType
			if isCategorized {
				markerTitle = tagName
			}
			request.markers = append(request.markers, sceneMarker{
				tag:   markerPrimaryTag,
				title: markerTitle,
				start: float64(tagReq.Start) / 1000,
			})
		}
	}

	return request
}
