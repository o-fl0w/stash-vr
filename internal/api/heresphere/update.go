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

type sceneUpdateInput struct {
	Rating       *int
	TagIds       *[]string
	StudioId     *string
	PerformerIds *[]string
}

func findOrCreateTag(ctx context.Context, client graphql.Client, name string) (string, error) {
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

func sceneUpdateInputFromReq(ctx context.Context, client graphql.Client, updateReq UpdateVideoData) sceneUpdateInput {
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
		for _, tagReq := range *updateReq.Tags {
			tagType, tagName, found := strings.Cut(tagReq.Name, ":")
			tagType = strings.ToLower(tagType)
			if !found {
				log.Ctx(ctx).Debug().Str("name", tagReq.Name).Msg("Tag not handled")
				continue
			}

			switch tagType {
			case legendTag, legendFull[legendTag]:
				id, err := findOrCreateTag(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("findOrCreateTag")
				}
				if input.TagIds == nil {
					input.TagIds = &[]string{}
				}
				*input.TagIds = append(*input.TagIds, id)
			case legendStudio, legendFull[legendStudio]:
				id, err := findOrCreateStudio(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("findOrCreateStudio")
				}
				input.StudioId = &id
			case legendPerformer, legendFull[legendPerformer]:
				id, err := findOrCreatePerformer(ctx, client, tagName)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("findOrCreatePerformer")
				}
				if input.PerformerIds == nil {
					input.PerformerIds = &[]string{}
				}
				*input.PerformerIds = append(*input.PerformerIds, id)
			default:
			}
		}
	}
	return input
}

func update(ctx context.Context, client graphql.Client, videoId string, updateReq UpdateVideoData) {
	input := sceneUpdateInputFromReq(ctx, client, updateReq)
	log.Ctx(ctx).Debug().Str("videoId", videoId).Interface("update req", updateReq).Interface("update input", input).Send()

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
