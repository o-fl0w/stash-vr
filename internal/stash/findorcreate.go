package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
)

func FindOrCreateTag(ctx context.Context, client graphql.Client, name string) (string, error) {
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

func FindOrCreateStudio(ctx context.Context, client graphql.Client, name string) (string, error) {
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

func FindOrCreatePerformer(ctx context.Context, client graphql.Client, name string) (string, error) {
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
