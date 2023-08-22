package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"net/url"
	"stash-vr/internal/efile"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
	"strconv"
)

func ESection(ctx context.Context, eSceneServer string, client graphql.Client) (section.Section, error) {
	eScenes, err := efile.GetList(eSceneServer)
	if err != nil {
		return section.Section{}, fmt.Errorf("esection: %w", err)
	}

	sceneIdSet := make(map[int]any)
	for _, e := range eScenes {
		id, _ := strconv.Atoi(e.SceneId)
		sceneIdSet[id] = struct{}{} // append(sceneIds, id) //[i] = id
	}
	sceneIds := make([]int, 0, len(sceneIdSet))
	for id, _ := range sceneIdSet {
		sceneIds = append(sceneIds, id)
	}

	response, err := gql.FindScenePreviewsByIds(ctx, client, sceneIds)
	if err != nil {
		return section.Section{}, err
	}

	s := section.Section{
		Name:     "EScenes",
		FilterId: efile.ESectionFilterId,
		Scenes:   make([]section.ScenePreview, len(eScenes)),
	}

	for i, e := range eScenes {
		for _, scene := range response.FindScenes.Scenes {
			if e.SceneId == scene.Id {
				es := e
				pp := section.ScenePreview{
					ScenePreviewParts: scene.ScenePreviewParts,
					EScene:            &es,
				}
				coverPath, _ := url.JoinPath(eSceneServer, "cover", fmt.Sprintf("%d_cover.png", e.Oshash))
				pp.ScenePreviewParts.Paths.Screenshot = coverPath

				s.Scenes[i] = pp
			}
		}
	}
	return s, nil
}
