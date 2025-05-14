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
		return "", fmt.Errorf("can not find or create with empty tag name")
	}
	findResponse, err := gql.FindTagByName(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("FindTagByName (%s): %w", name, err)
	}
	if len(findResponse.FindTags.Tags) == 0 {
		createResponse, err := gql.TagCreate(ctx, client, name)
		if err != nil {
			return "", fmt.Errorf("TagCreate (%s): %w", name, err)
		}
		log.Ctx(ctx).Info().Str("name", name).Str("id", createResponse.TagCreate.Id).Msg("Tag created in stash")
		return createResponse.TagCreate.Id, nil
	}
	return findResponse.FindTags.Tags[0].Id, nil
}
