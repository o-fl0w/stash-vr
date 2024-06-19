package deovr

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/heatmap"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"strings"
)

type videoData struct {
	Authorized     string `json:"authorized"`
	FullAccess     bool   `json:"fullAccess"`
	Title          string `json:"title"`
	Id             string `json:"id"`
	VideoLength    int    `json:"videoLength"`
	Is3d           bool   `json:"is3d"`
	ScreenType     string `json:"screenType"`
	StereoMode     string `json:"stereoMode"`
	SkipIntro      int    `json:"skipIntro"`
	VideoThumbnail string `json:"videoThumbnail,omitempty"`
	VideoPreview   string `json:"videoPreview,omitempty"`
	ThumbnailUrl   string `json:"thumbnailUrl"`

	Subtitles []subtitle `json:"subtitles"`

	TimeStamps []timeStamp `json:"timeStamps,omitempty"`

	Encodings []encoding `json:"encodings"`
}

type timeStamp struct {
	Ts   int    `json:"ts"`
	Name string `json:"name"`
}

type subtitle struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type encoding struct {
	Name         string        `json:"name"`
	VideoSources []videoSource `json:"videoSources"`
}

type videoSource struct {
	Resolution int    `json:"resolution"`
	Url        string `json:"url"`
}

func buildVideoData(ctx context.Context, client graphql.Client, baseUrl string, sceneId string) (videoData, error) {
	findSceneResponse, err := gql.FindSceneFull(ctx, client, sceneId)
	if err != nil {
		return videoData{}, fmt.Errorf("FindScene: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return videoData{}, fmt.Errorf("FindScene: not found")
	}
	s := findSceneResponse.FindScene.SceneFullParts

	if len(s.SceneScanParts.Files) == 0 {
		return videoData{}, fmt.Errorf("scene %s has no files", sceneId)
	}

	thumbnailUrl := stash.ApiKeyed(s.Paths.Screenshot)
	if !config.Get().IsHeatmapDisabled && s.ScriptParts.Interactive && s.ScriptParts.Paths.Interactive_heatmap != "" {
		thumbnailUrl = heatmap.GetCoverUrl(baseUrl, sceneId)
	}

	title := s.Title
	if title == "" {
		title = s.SceneScanParts.Files[0].Basename
	}

	vd := videoData{
		Authorized:   "1",
		FullAccess:   true,
		Title:        title,
		Id:           s.Id,
		VideoLength:  int(s.SceneScanParts.Files[0].Duration),
		SkipIntro:    0,
		VideoPreview: stash.ApiKeyed(s.Paths.Preview),
		ThumbnailUrl: thumbnailUrl,
	}

	setStreamSources(ctx, s, &vd)
	setSubtitles(s, &vd)
	setMarkers(s, &vd)
	set3DFormat(s, &vd)

	return vd, nil
}

func setSubtitles(s gql.SceneFullParts, videoData *videoData) {
	if s.Captions != nil {
		for _, c := range s.Captions {
			videoData.Subtitles = append(videoData.Subtitles, subtitle{
				Title: fmt.Sprintf(".%s.%s", c.Language_code, c.Caption_type),
				Url:   stash.ApiKeyed(fmt.Sprintf("%s?lang=%s&type=%s", s.Paths.Caption, c.Language_code, c.Caption_type)),
			})
		}
	}
}

func setStreamSources(ctx context.Context, s gql.SceneFullParts, videoData *videoData) {
	streams := stash.GetStreams(ctx, s.StreamsParts, false)
	videoData.Encodings = make([]encoding, len(streams))
	for i, stream := range streams {
		videoData.Encodings[i] = encoding{
			Name:         stream.Name,
			VideoSources: make([]videoSource, len(stream.Sources)),
		}
		for j, source := range stream.Sources {
			videoData.Encodings[i].VideoSources[j] = videoSource{
				Resolution: source.Resolution,
				Url:        source.Url,
			}
		}
	}
}

func setMarkers(s gql.SceneFullParts, videoData *videoData) {
	for _, sm := range s.Scene_markers {
		sb := strings.Builder{}
		sb.WriteString(sm.Primary_tag.Name)
		if sm.Title != "" {
			sb.WriteString(":")
			sb.WriteString(sm.Title)
		}
		ts := timeStamp{
			Ts:   int(sm.Seconds),
			Name: sb.String(),
		}
		videoData.TimeStamps = append(videoData.TimeStamps, ts)
	}
}

func set3DFormat(s gql.SceneFullParts, videoData *videoData) {
	for _, tag := range s.Tags {
		switch tag.Name {
		case "DOME":
			videoData.Is3d = true
			videoData.ScreenType = "dome"
			videoData.StereoMode = "sbs"
			continue
		case "SPHERE":
			videoData.Is3d = true
			videoData.ScreenType = "sphere"
			videoData.StereoMode = "sbs"
			continue
		case "FISHEYE":
			videoData.Is3d = true
			videoData.ScreenType = "fisheye"
			videoData.StereoMode = "sbs"
			continue
		case "MKX200":
			videoData.Is3d = true
			videoData.ScreenType = "mkx200"
			videoData.StereoMode = "sbs"
			continue
		case "RF52":
			videoData.Is3d = true
			videoData.ScreenType = "rf52"
			videoData.StereoMode = "sbs"
			continue
		case "SBS":
			videoData.Is3d = true
			videoData.StereoMode = "sbs"
			continue
		case "TB":
			videoData.Is3d = true
			videoData.StereoMode = "tb"
			continue
		}
	}
}
