package heresphere

import (
	"context"
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/sections"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/stimhub"
	"stash-vr/internal/util"
	"strconv"
	"time"
)

type scanDoc struct {
	ScanData []scanData `json:"scanData"`
}

type scanData struct {
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

func buildScan(ctx context.Context, stashClient graphql.Client, stimhubClient *stimhub.Client, baseUrl string) (scanDoc, error) {
	ss := sections.Get(ctx, stashClient, stimhubClient)
	sceneIdMap := make(map[int]any)

	var stimSection *section.Section

	for _, s := range ss {
		if s.FilterId == stimhub.FilterId {
			stimSection = &s
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
	response, err := gql.FindSceneScansByIds(ctx, stashClient, sceneIds)
	if err != nil {
		return scanDoc{}, fmt.Errorf("FindSceneScansByIds: %w", err)
	}

	sceneScans := util.Transform[*gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene, scanData](
		func(scene *gql.FindSceneScansByIdsFindScenesFindScenesResultTypeScenesScene) (scanData, error) {
			return mkScanData(scene.SceneScanParts, baseUrl), nil
		}).Ordered(response.FindScenes.Scenes)

	if stimSection != nil {
		eSceneScans := util.Transform[section.ScenePreview, scanData](
			func(p section.ScenePreview) (scanData, error) {
				for _, el := range sceneScans {
					if p.ScenePreviewParts.Id == el.id {
						stimScene := stimhub.Get(p.StimAudioCrc32, p.GetId())
						eel := el
						eel.Title = stimScene.Title
						eel.DateAdded = stimScene.DateAdded.Format(time.DateOnly)
						eel.Link = getVideoDataUrl(baseUrl, stimhub.MakeStimSceneId(p.GetId(), p.StimAudioCrc32))
						eel.Tags = make([]tag, len(el.Tags))
						copy(eel.Tags, el.Tags)
						eel.Tags = append(eel.Tags, getStimSceneTags(*stimScene)...)
						return eel, nil
					}
				}
				return scanData{}, errors.New("scandataelement not found")
			}).Ordered(stimSection.Scenes)
		sceneScans = append(sceneScans, eSceneScans...)
	}

	log.Ctx(ctx).Trace().Int("count", len(sceneScans)).Msg("/scan")
	return scanDoc{ScanData: sceneScans}, nil
}

func mkScanData(sp gql.SceneScanParts, baseUrl string) scanData {
	return scanData{
		id:           sp.Id,
		Link:         getVideoDataUrl(baseUrl, sp.Id),
		Title:        util.FirstNonEmpty(sp.Title, sp.GetFiles()[0].Basename),
		DateReleased: sp.Date,
		DateAdded:    sp.Created_at.Format(time.DateOnly),
		Duration:     sp.Files[0].Duration,
		Rating:       float32(sp.Rating100) / 20,
		Favorites:    sp.O_counter,
		IsFavorite:   ContainsFavoriteTag(sp.TagPartsArray),
		Tags:         getTags(sp),
	}
}
