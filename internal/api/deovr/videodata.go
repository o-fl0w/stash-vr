package deovr

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
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
	Categories []Category  `json:"categories,omitempty"`
	Actors     []Tag       `json:"actors"`

	Encodings []Encoding `json:"encodings"`
}

type Category struct {
	Tag Tag `json:"tag"`
}

type Tag struct {
	Id   string `json:"id"`
	Name string `json:"name"`
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
	setTags(s, &videoData)
	setStudios(s, &videoData)
	setMarkers(s, &videoData)
	setPerformers(s, &videoData)
	set3DFormat(s, &videoData)

	return videoData, nil
}

func setStreamSources(ctx context.Context, s gql.FullSceneParts, videoData *VideoData) {
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

func setTags(s gql.FullSceneParts, videoData *VideoData) {
	for _, tag := range s.Tags {
		videoData.Categories = append(videoData.Categories, Category{Tag{
			Id:   tag.Id,
			Name: fmt.Sprintf("#:%s", tag.Name),
		}})
	}
}

func setStudios(s gql.FullSceneParts, videoData *VideoData) {
	if s.Studio != nil {
		videoData.Categories = append(videoData.Categories, Category{Tag{
			Id:   s.Studio.Id,
			Name: fmt.Sprintf("Studio:%s", s.Studio.Name),
		}})
	}
}

func setMarkers(s gql.FullSceneParts, videoData *VideoData) {
	for _, sm := range s.Scene_markers {
		ts := TimeStamp{
			Ts:   int(sm.Seconds),
			Name: sm.Title,
		}
		videoData.TimeStamps = append(videoData.TimeStamps, ts)
	}
}

func setPerformers(s gql.FullSceneParts, videoData *VideoData) {
	for _, p := range s.Performers {
		t := Tag{
			Id:   p.Id,
			Name: p.Name,
		}
		videoData.Actors = append(videoData.Actors, t)
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
