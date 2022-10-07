package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/cache"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strconv"
)

type scanDoc struct {
	ScanData []scanDataElement `json:"scanData"`
}

type scanDataElement struct {
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

func buildScan(ctx context.Context, client graphql.Client, baseUrl string) (scanDoc, error) {
	sections := cache.GetSections(ctx, client)
	sceneIdMap := make(map[int]any)
	for _, s := range sections {
		for _, preview := range s.PreviewPartsList {
			id, _ := strconv.Atoi(preview.Id)
			sceneIdMap[id] = struct{}{}
		}
	}
	sceneIds := make([]int, 0, len(sceneIdMap))
	for id := range sceneIdMap {
		sceneIds = append(sceneIds, id)
	}
	response, err := gql.FindSceneScansByIds(ctx, client, sceneIds)
	if err != nil {
		return scanDoc{}, fmt.Errorf("FindSceneScansByIds: %w", err)
	}

	sceneScans := util.Transform[*gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene, scanDataElement](
		func(part *gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene) *scanDataElement {
			return &scanDataElement{
				Link:         getVideoDataUrl(baseUrl, part.Id),
				Title:        part.Title,
				DateReleased: part.Date,
				DateAdded:    part.Created_at.Format("2006-01-02"),
				Duration:     part.File.Duration,
				Rating:       float32(part.Rating),
				Favorites:    part.O_counter,
				IsFavorite:   ContainsFavoriteTag(part.TagPartsArray),
				Tags:         getTags(part.SceneScanParts),
			}
		}).Ordered(response.FindScenes.Scenes)
	return scanDoc{ScanData: sceneScans}, nil
}
