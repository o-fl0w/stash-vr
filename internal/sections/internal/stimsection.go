package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/stimhub"
	"strconv"
)

func StimSection(ctx context.Context, stimhubClient stimhub.Client, stashClient graphql.Client) (section.Section, error) {
	stimScenes, err := stimhub.StimScenes(ctx, stimhubClient)
	if err != nil {
		return section.Section{}, fmt.Errorf("stimsection: %w", err)
	}
	stimhub.Set(stimScenes)

	sceneIdSet := make(map[int]any)
	for _, e := range stimScenes {
		id, _ := strconv.Atoi(e.SceneId)
		sceneIdSet[id] = struct{}{} // append(sceneIds, id) //[i] = ideee
	}
	sceneIds := make([]int, 0, len(sceneIdSet))
	for id := range sceneIdSet {
		sceneIds = append(sceneIds, id)
	}

	response, err := gql.FindScenePreviewsByIds(ctx, stashClient, sceneIds)
	if err != nil {
		return section.Section{}, err
	}

	s := section.Section{
		Name:     "Stim",
		FilterId: stimhub.FilterId,
		Scenes:   make([]section.ScenePreview, len(stimScenes)),
	}

	for i, stimScene := range stimScenes {
		for _, stashScene := range response.FindScenes.Scenes {
			if stimScene.SceneId == stashScene.Id {
				ss := stimScene
				preview := section.ScenePreview{
					ScenePreviewParts: stashScene.ScenePreviewParts,
					StimAudioCrc32:    ss.AudioCrc32,
				}
				preview.ScenePreviewParts.Paths.Screenshot = stimhubClient.ThumbnailUrl(ss.AudioCrc32)

				s.Scenes[i] = preview
			}
		}
	}
	return s, nil
}
