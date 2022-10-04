package stash

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/stash/gql"
)

func SceneToggleOrganized(ctx context.Context, client graphql.Client, id string) (bool, error) {
	isOrganizedResponse, err := gql.IsSceneOrganized(ctx, client, id)
	if err != nil {
		return false, fmt.Errorf("IsSceneOrganized: %w", err)
	}
	if isOrganizedResponse.FindScene == nil {
		return false, fmt.Errorf("IsSceneOrganized: Scene %s not found", id)
	}
	updateResponse, err := gql.SceneUpdateOrganized(ctx, client, id, !isOrganizedResponse.FindScene.Organized)
	if err != nil {
		return false, fmt.Errorf("SceneUpdateOrganized: %w", err)
	}
	return updateResponse.SceneUpdate.Organized, nil
}
