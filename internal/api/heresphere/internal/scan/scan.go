package scan

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/heresphere/internal/index"
	"stash-vr/internal/api/heresphere/internal/tag"
	"stash-vr/internal/api/heresphere/internal/videodata"
	"stash-vr/internal/section"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strconv"
)

type Document struct {
	ScanData []ScanDataElement `json:"scanData"`
}

type ScanDataElement struct {
	Link         index.VideoDataUrl `json:"link"`
	Title        string             `json:"title"`
	DateReleased string             `json:"dateReleased"`
	DateAdded    string             `json:"dateAdded"`
	Duration     float64            `json:"duration"`
	Rating       float32            `json:"rating"`
	Favorites    int                `json:"favorites"`
	IsFavorite   bool               `json:"isFavorite"`
	Tags         []tag.Tag          `json:"tags"`
}

func Build(ctx context.Context, client graphql.Client, baseUrl string) (Document, error) {
	sections := section.Get(ctx, client)
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
		return Document{}, fmt.Errorf("FindSceneScansByIds: %w", err)
	}

	sceneScans := util.Transform[*gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene, ScanDataElement](
		func(part *gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene) *ScanDataElement {
			return &ScanDataElement{
				Link:         index.GetVideoDataUrl(baseUrl, part.Id),
				Title:        part.Title,
				DateReleased: part.Date,
				DateAdded:    part.Created_at.Format("2006-01-02"),
				Duration:     part.File.Duration,
				Rating:       float32(part.Rating),
				Favorites:    part.O_counter,
				IsFavorite:   videodata.ContainsFavoriteTag(part.TagPartsArray),
				Tags:         tag.GetTags(part.SceneDetailsParts),
			}
		}).Ordered(response.FindScenes.Scenes)
	return Document{ScanData: sceneScans}, nil
}
