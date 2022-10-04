package deovr

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"strings"
)

type VideoData struct {
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

	TimeStamps []TimeStamp `json:"timeStamps,omitempty"`

	Encodings []Encoding `json:"encodings"`
}

type TimeStamp struct {
	Ts   int    `json:"ts"`
	Name string `json:"name"`
}

type Encoding struct {
	Name         string        `json:"name"`
	VideoSources []VideoSource `json:"videoSources"`
}

type VideoSource struct {
	Resolution int    `json:"resolution"`
	Url        string `json:"url"`
}

func buildVideoData(ctx context.Context, client graphql.Client, sceneId string) (VideoData, error) {
	findSceneResponse, err := gql.FindScene(ctx, client, sceneId)
	if err != nil {
		return VideoData{}, fmt.Errorf("FindScene: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return VideoData{}, fmt.Errorf("FindScene: not found")
	}
	s := findSceneResponse.FindScene.FullSceneParts

	videoData := VideoData{
		Authorized:   "1",
		FullAccess:   true,
		Title:        s.Title,
		Id:           s.Id,
		VideoLength:  int(s.File.Duration),
		SkipIntro:    0,
		VideoPreview: stash.ApiKeyed(s.Paths.Preview),
		ThumbnailUrl: stash.ApiKeyed(s.Paths.Screenshot),
	}

	setStreamSources(ctx, s, &videoData)
	setMarkers(s, &videoData)
	set3DFormat(s, &videoData)

	return videoData, nil
}

func setStreamSources(ctx context.Context, s gql.FullSceneParts, videoData *VideoData) {
	log.Ctx(ctx).Trace().Str("codec", s.File.Video_codec).Send()
	streams := stash.GetStreams(ctx, s, false)
	videoData.Encodings = make([]Encoding, len(streams))
	for i, stream := range streams {
		videoData.Encodings[i] = Encoding{
			Name:         stream.Name,
			VideoSources: make([]VideoSource, len(stream.Sources)),
		}
		for j, source := range stream.Sources {
			videoData.Encodings[i].VideoSources[j] = VideoSource{
				Resolution: source.Resolution,
				Url:        source.Url,
			}
		}
	}
}

func setMarkers(s gql.FullSceneParts, videoData *VideoData) {
	for _, sm := range s.Scene_markers {
		sb := strings.Builder{}
		sb.WriteString(sm.Primary_tag.Name)
		if sm.Title != "" {
			sb.WriteString(":")
			sb.WriteString(sm.Title)
		}
		ts := TimeStamp{
			Ts:   int(sm.Seconds),
			Name: sb.String(),
		}
		videoData.TimeStamps = append(videoData.TimeStamps, ts)
	}
}

func set3DFormat(s gql.FullSceneParts, videoData *VideoData) {
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
