package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"math"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
)

type UpdateVideoData struct {
	Rating     *float32 `json:"rating,omitempty"`
	Tags       *[]Tag   `json:"tags,omitempty"`
	DeleteFile *bool    `json:"deleteFile,omitempty"`
}

func (v UpdateVideoData) IsUpdateRequest() bool {
	return v.Rating != nil || v.Tags != nil
}

func (v UpdateVideoData) IsDeleteRequest() bool {
	return v.DeleteFile != nil && *v.DeleteFile
}

type sceneUpdateInput struct {
	Rating       *int
	TagIds       *[]string
	StudioId     *string
	PerformerIds *[]string
}

type SceneMarker struct {
	tag   string
	title string
	start float64
}

func setSceneMarkers(ctx context.Context, client graphql.Client, sceneId string, markers []SceneMarker) error {
	response, err := gql.FindSceneMarkers(ctx, client, sceneId)
	if err != nil {
		return fmt.Errorf("FindSceneMarkers: %w", err)
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
	return nil
}

func sceneUpdateInputFromReq(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) sceneUpdateInput {
	input := sceneUpdateInput{}
	if updateReq.Rating != nil {
		if *updateReq.Rating == 0.5 {
			//special case to set zero rating
			input.Rating = util.Ptr(0)
		} else {
			input.Rating = util.Ptr(int(math.Ceil(float64(*updateReq.Rating))))
		}
	}
	if updateReq.Tags != nil {
		var sceneMarkers []SceneMarker
		for _, tagReq := range *updateReq.Tags {
			tagType, tagName, isCategorized := strings.Cut(tagReq.Name, ":")

			if isCategorized && legendTag.IsMatch(tagType) {
				id, err := stash.FindOrCreateTag(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("FindOrCreateTag")
				}
				if input.TagIds == nil {
					input.TagIds = &[]string{}
				}
				*input.TagIds = append(*input.TagIds, id)
			} else if isCategorized && legendStudio.IsMatch(tagType) {
				id, err := stash.FindOrCreateStudio(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("FindOrCreateStudio")
				}
				input.StudioId = &id
			} else if isCategorized && legendPerformer.IsMatch(tagType) {
				id, err := stash.FindOrCreatePerformer(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("FindOrCreatePerformer")
				}
				if input.PerformerIds == nil {
					input.PerformerIds = &[]string{}
				}
				*input.PerformerIds = append(*input.PerformerIds, id)
			} else {
				var markerTitle string

				markerPrimaryTag := tagType
				if isCategorized {
					markerTitle = tagName
				}
				log.Ctx(ctx).Debug().Str("primary tag", markerPrimaryTag).Str("title", markerTitle).Msg("Scene marker")
				sceneMarkers = append(sceneMarkers, SceneMarker{
					tag:   markerPrimaryTag,
					title: markerTitle,
					start: float64(tagReq.Start) / 1000,
				})
			}
		}
		if len(sceneMarkers) != 0 {
			err := setSceneMarkers(ctx, client, sceneId, sceneMarkers)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to set scene markers")
			}
		}
	}
	return input
}

func update(ctx context.Context, client graphql.Client, sceneId string, updateReq UpdateVideoData) {
	input := sceneUpdateInputFromReq(ctx, client, sceneId, updateReq)
	log.Ctx(ctx).Trace().Str("sceneId", sceneId).Interface("update req", updateReq).Interface("update input", input).Send()

	if input.Rating != nil {
		_, err := gql.SceneUpdateRating(ctx, client, sceneId, *input.Rating)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateRating")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Int("new rating", *input.Rating).Msg("Rating updated")
	}
	if input.TagIds != nil {
		_, err := gql.SceneUpdateTags(ctx, client, sceneId, *input.TagIds)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateTags")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Interface("new tags", *input.TagIds).Msg("Tags updated")
	}
	if input.StudioId != nil {
		_, err := gql.SceneUpdateStudio(ctx, client, sceneId, *input.StudioId)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateStudio")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Interface("new studio", *input.StudioId).Msg("Studio updated")
	}
	if input.PerformerIds != nil {
		_, err := gql.SceneUpdatePerformers(ctx, client, sceneId, *input.PerformerIds)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdatePerformers")
		}
		log.Ctx(ctx).Debug().Str("sceneId", sceneId).Interface("new performers", *input.TagIds).Msg("Performers updated")
	}
}

func destroy(ctx context.Context, client graphql.Client, sceneId string) error {
	_, err := gql.SceneDestroy(ctx, client, sceneId)
	if err != nil {
		return fmt.Errorf("SceneDestroy: %w", err)
	}
	log.Ctx(ctx).Debug().Str("sceneId", sceneId).Msg("Requested stash to delete scene, file and generated content")
	return nil
}
