package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/efile"
	"stash-vr/internal/sections"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/title"
	"stash-vr/internal/util"
	"strconv"
)

type scanDoc struct {
	ScanData []scanDataElement `json:"scanData"`
}

type scanDataElement struct {
	id           string
	Link         string  `json:"link"`
	Title        string  `json:"title"`
	DateReleased string  `json:"dateReleased"`
	DateAdded    string  `json:"dateAdded"`
	Duration     float64 `json:"duration"`
	Rating       float32 `json:"rating"`
	Favorites    int     `json:"favorites"`
	IsFavorite   bool    `json:"isFavorite"`
	Tags         []tag   `json:"tags"`
}

type eScene struct {
	sceneId     string
	eFileSuffix string
}

func buildScan(ctx context.Context, client graphql.Client, baseUrl string) (scanDoc, error) {
	ss := sections.Get(ctx, client)
	sceneIdMap := make(map[int]any)
	var eScenes []eScene
	for _, s := range ss {
		for _, preview := range s.Scene {
			sceneId, eFileSuffix, isEScene := efile.GetSceneIdAndEFileSuffix(preview.Id)

			if isEScene {
				eScenes = append(eScenes, eScene{
					sceneId:     sceneId,
					eFileSuffix: eFileSuffix,
				})
				continue
			}

			sid, _ := strconv.Atoi(sceneId)
			sceneIdMap[sid] = struct{}{}
		}
	}
	sceneIds := make([]int, 0, len(sceneIdMap))
	for sceneId := range sceneIdMap {
		sceneIds = append(sceneIds, sceneId)
	}
	response, err := gql.FindSceneScansByIds(ctx, client, sceneIds)
	if err != nil {
		return scanDoc{}, fmt.Errorf("FindSceneScansByIds: %w", err)
	}

	sceneScans := util.Transform[*gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene, scanDataElement](
		func(scene *gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene) (scanDataElement, error) {
			return scanDataElement{
				id:           scene.Id,
				Link:         getVideoDataUrl(baseUrl, scene.Id),
				Title:        title.GetSceneTitle(scene.Title, scene.GetFiles()[0].Basename),
				DateReleased: scene.Date,
				DateAdded:    scene.Created_at.Format("2006-01-02"),
				Duration:     scene.Files[0].Duration,
				Rating:       float32(scene.Rating100) / 20,
				Favorites:    scene.O_counter,
				IsFavorite:   ContainsFavoriteTag(scene.TagPartsArray),
				Tags:         getTags(scene.SceneScanParts),
			}, nil
		}).Ordered(response.FindScenes.Scenes)
	for _, e := range eScenes {
		el := findScanDataElement(e.sceneId, sceneScans)
		el.Link = getVideoDataUrl(baseUrl, efile.MakeESceneIdWithEFileSuffix(e.sceneId, e.eFileSuffix))
		el.Title = efile.MakeESceneTitleWithEFileSuffix(el.Title, e.eFileSuffix)
		sceneScans = append(sceneScans, el)
	}
	log.Ctx(ctx).Trace().Int("count", len(sceneScans)).Msg("/scan")
	return scanDoc{ScanData: sceneScans}, nil
}

func findScanDataElement(sceneId string, es []scanDataElement) scanDataElement {
	for _, e := range es {
		if e.id == sceneId {
			return e
		}
	}
	return scanDataElement{}
}
