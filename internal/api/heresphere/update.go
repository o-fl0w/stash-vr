package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"math"
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

func findOrCreateTag(ctx context.Context, client graphql.Client, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty tag name")
	}
	findResponse, err := gql.FindTagByName(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("FindTagByName name='%s': %w", name, err)
	}
	if len(findResponse.FindTags.Tags) == 0 {
		log.Ctx(ctx).Debug().Str("name", name).Msg("Create tag")
		createResponse, err := gql.TagCreate(ctx, client, name)
		if err != nil {
			return "", fmt.Errorf("TagCreate name='%s': %w", name, err)
		}
		return createResponse.TagCreate.Id, nil
	} else {
		return findResponse.FindTags.Tags[0].Id, nil
	}
}

func findOrCreateStudio(ctx context.Context, client graphql.Client, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty studio name")
	}
	findResponse, err := gql.FindStudioByName(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("FindStudioByName name='%s': %w", name, err)
	}
	if len(findResponse.FindStudios.Studios) == 0 {
		log.Ctx(ctx).Debug().Str("name", name).Msg("Create tag!")
		createResponse, err := gql.StudioCreate(ctx, client, name)
		if err != nil {
			return "", fmt.Errorf("StudioCreate name='%s': %w", name, err)
		}
		return createResponse.StudioCreate.Id, nil
	} else {
		return findResponse.FindStudios.Studios[0].Id, nil
	}
}

func findOrCreatePerformer(ctx context.Context, client graphql.Client, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty performer name")
	}
	findResponse, err := gql.FindPerformerByName(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("FindPerformerByName name='%s': %w", name, err)
	}
	if len(findResponse.FindPerformers.Performers) == 0 {
		log.Ctx(ctx).Debug().Str("name", name).Msg("Create performer")
		createResponse, err := gql.PerformerCreate(ctx, client, name)
		if err != nil {
			return "", fmt.Errorf("PerformerCreate name='%s': %w", name, err)
		}
		return createResponse.PerformerCreate.Id, nil
	} else {
		return findResponse.FindPerformers.Performers[0].Id, nil
	}
}

type SceneMarker struct {
	tag   string
	title string
	start float64
}

func setSceneMarkers(ctx context.Context, client graphql.Client, videoId string, markers []SceneMarker) error {
	response, err := gql.FindSceneMarkers(ctx, client, videoId)
	if err != nil {
		return fmt.Errorf("FindSceneMarkers: %w", err)
	}
	for _, smt := range response.SceneMarkerTags {
		for _, sm := range smt.Scene_markers {
			if _, err := gql.SceneMarkerDestroy(ctx, client, sm.Id); err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("videoId", videoId).Str("sceneMarkerId", sm.Id).Str("sceneMarkerTitle", sm.Title).Msg("Failed to delete scene marker")
				continue
			}
		}
	}
	for _, m := range markers {
		tagId, err := findOrCreateTag(ctx, client, m.tag)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("videoId", videoId).Msg("findOrCreateTag")
			continue
		}
		_, err = gql.SceneMarkerCreate(ctx, client, videoId, tagId, m.start, m.title)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("videoId", videoId).Interface("marker", m).Msg("SceneMarkerCreate")
			continue
		}
	}
	return nil
}

func sceneUpdateInputFromReq(ctx context.Context, client graphql.Client, videoId string, updateReq UpdateVideoData) sceneUpdateInput {
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
				id, err := findOrCreateTag(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("findOrCreateTag")
				}
				if input.TagIds == nil {
					input.TagIds = &[]string{}
				}
				*input.TagIds = append(*input.TagIds, id)
			} else if isCategorized && legendStudio.IsMatch(tagType) {
				id, err := findOrCreateStudio(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("findOrCreateStudio")
				}
				input.StudioId = &id
			} else if isCategorized && legendPerformer.IsMatch(tagType) {
				id, err := findOrCreatePerformer(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("findOrCreatePerformer")
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
			err := setSceneMarkers(ctx, client, videoId, sceneMarkers)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Failed to set scene markers")
			}
		}
	}
	return input
}

func update(ctx context.Context, client graphql.Client, videoId string, updateReq UpdateVideoData) {
	input := sceneUpdateInputFromReq(ctx, client, videoId, updateReq)
	log.Ctx(ctx).Trace().Str("videoId", videoId).Interface("update req", updateReq).Interface("update input", input).Send()

	if input.Rating != nil {
		_, err := gql.SceneUpdateRating(ctx, client, videoId, *input.Rating)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateRating")
		}
		log.Ctx(ctx).Debug().Str("videoId", videoId).Int("new rating", *input.Rating).Msg("Rating updated")
	}
	if input.TagIds != nil {
		_, err := gql.SceneUpdateTags(ctx, client, videoId, *input.TagIds)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateTags")
		}
		log.Ctx(ctx).Debug().Str("videoId", videoId).Interface("new tags", *input.TagIds).Msg("Tags updated")
	}
	if input.StudioId != nil {
		_, err := gql.SceneUpdateStudio(ctx, client, videoId, *input.StudioId)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdateStudio")
		}
		log.Ctx(ctx).Debug().Str("videoId", videoId).Interface("new studio", *input.StudioId).Msg("Studio updated")
	}
	if input.PerformerIds != nil {
		_, err := gql.SceneUpdatePerformers(ctx, client, videoId, *input.PerformerIds)
		if err != nil {
			log.Ctx(ctx).Warn().Interface("input", input).Msg("SceneUpdatePerformers")
		}
		log.Ctx(ctx).Debug().Str("videoId", videoId).Interface("new performers", *input.TagIds).Msg("Performers updated")
	}
}

func destroy(ctx context.Context, client graphql.Client, videoId string) error {
	_, err := gql.SceneDestroy(ctx, client, videoId)
	if err != nil {
		return fmt.Errorf("SceneDestroy: %w", err)
	}
	log.Ctx(ctx).Debug().Str("videoId", videoId).Msg("Requested stash to delete scene, file and generated content")
	return nil
}
