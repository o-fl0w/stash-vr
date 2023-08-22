package heresphere

import (
	"context"
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/efile"
	"stash-vr/internal/sections"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/title"
	"stash-vr/internal/util"
	"strconv"
	"time"
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

func buildScan(ctx context.Context, client graphql.Client, baseUrl string) (scanDoc, error) {
	ss := sections.Get(ctx, client)
	sceneIdMap := make(map[int]any)

	var eSection *section.Section

	for _, s := range ss {
		if s.FilterId == efile.ESectionFilterId {
			eSection = &s
		}
		for _, preview := range s.Scenes {
			sid, _ := strconv.Atoi(preview.ScenePreviewParts.Id)
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
			return mkScanDataElement(scene.SceneScanParts, baseUrl), nil
		}).Ordered(response.FindScenes.Scenes)

	if eSection != nil {
		eSceneScans := util.Transform[section.ScenePreview, scanDataElement](
			func(p section.ScenePreview) (scanDataElement, error) {
				for _, el := range sceneScans {
					if p.ScenePreviewParts.Id == el.id {
						eel := el
						eel.Title = p.EScene.Title
						eel.DateAdded = p.EScene.AddedTime.Format(time.DateOnly)
						eel.Link = getVideoDataUrl(baseUrl, p.Id())
						eel.Tags = make([]tag, len(el.Tags))
						copy(eel.Tags, el.Tags)
						eel.Tags = append(eel.Tags, getETags(*p.EScene)...)
						return eel, nil
					}
				}
				return scanDataElement{}, errors.New("scandataelement not found")
			}).Ordered(eSection.Scenes)
		sceneScans = append(sceneScans, eSceneScans...)
	}

	log.Ctx(ctx).Trace().Int("count", len(sceneScans)).Msg("/scan")
	return scanDoc{ScanData: sceneScans}, nil
}

func mkScanDataElement(sp gql.SceneScanParts, baseUrl string) scanDataElement {
	return scanDataElement{
		id:           sp.Id,
		Link:         getVideoDataUrl(baseUrl, sp.Id),
		Title:        title.GetSceneTitle(sp.Title, sp.GetFiles()[0].Basename),
		DateReleased: sp.Date,
		DateAdded:    sp.Created_at.Format(time.DateOnly),
		Duration:     sp.Files[0].Duration,
		Rating:       float32(sp.Rating100) / 20,
		Favorites:    sp.O_counter,
		IsFavorite:   ContainsFavoriteTag(sp.TagPartsArray),
		Tags:         getTags(sp),
	}
}
