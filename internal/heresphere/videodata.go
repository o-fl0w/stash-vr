package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

type VideoData struct {
	Access         int      `json:"access"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	DateAdded      string   `json:"dateAdded"`
	ThumbnailImage string   `json:"thumbnailImage"`
	ThumbnailVideo string   `json:"thumbnailVideo"`
	Duration       int      `json:"duration"`
	Rating         float32  `json:"rating"`
	Media          []Media  `json:"media"`
	Tags           []Tag    `json:"tags"`
	Projection     string   `json:"projection"`
	Stereo         string   `json:"stereo"`
	Lens           string   `json:"lens"`
	Fov            float32  `json:"fov"`
	Scripts        []Script `json:"scripts"`
}

type Tag struct {
	Name   string  `json:"name"`
	Start  int     `json:"start"`
	End    int     `json:"end"`
	Track  *int    `json:"track,omitempty"`
	Rating float32 `json:"rating"`
}

type Media struct {
	Name    string   `json:"name"`
	Sources []Source `json:"sources"`
}

type Source struct {
	Resolution int    `json:"resolution"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Size       int    `json:"size"`
	Url        string `json:"url"`
}

type Script struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func buildVideoData(ctx context.Context, client graphql.Client, videoId string) (VideoData, error) {
	findSceneResponse, err := gql.FindScene(ctx, client, videoId)
	if err != nil {
		return VideoData{}, fmt.Errorf("FindScene: %w", err)
	}
	s := findSceneResponse.FindScene.FullSceneParts

	videoData := VideoData{
		Access:         1,
		Title:          s.Title,
		Description:    s.Details,
		DateAdded:      s.Created_at.Format("2006-01-02"),
		ThumbnailImage: stash.ApiKeyed(s.Paths.Screenshot),
		ThumbnailVideo: stash.ApiKeyed(s.Paths.Preview),
		Duration:       int(s.File.Duration) * 1000,
		Rating:         float32(s.Rating),
	}

	setStreamSources(s, &videoData)
	set3DFormat(s, &videoData)
	setTags(s, &videoData)
	setStudios(s, &videoData)
	setMarkers(s, &videoData)
	setPerformers(s, &videoData)
	setScripts(s, &videoData)

	return videoData, nil
}

func setScripts(s gql.FullSceneParts, videoData *VideoData) {
	if s.ScriptParts.Interactive {
		videoData.Scripts = append(videoData.Scripts, Script{
			Name: fmt.Sprintf("Script-%s", s.Title),
			Url:  s.ScriptParts.Paths.Funscript,
		})
	}
}

func setPerformers(s gql.FullSceneParts, videoData *VideoData) {
	for _, p := range s.Performers {
		t := Tag{
			Name:   fmt.Sprintf("Performer:%s", p.Name),
			Start:  0,
			End:    0,
			Track:  util.Ptr(0),
			Rating: float32(p.Rating),
		}
		videoData.Tags = append(videoData.Tags, t)
	}
}

func setMarkers(s gql.FullSceneParts, videoData *VideoData) {
	for i, sm := range s.Scene_markers {
		t := Tag{
			Name:  fmt.Sprintf("@:%s", sm.Title),
			Start: int(sm.Seconds * 1000),
			End:   0,
			Track: util.Ptr(1 + i),
		}
		videoData.Tags = append(videoData.Tags, t)
	}
}

func setStudios(s gql.FullSceneParts, videoData *VideoData) {
	if s.Studio != nil {
		t := Tag{
			Name:   fmt.Sprintf("Studio:%s", s.Studio.Name),
			Rating: float32(s.Studio.Rating),
			Track:  util.Ptr(0),
		}
		videoData.Tags = append(videoData.Tags, t)
	}
}

func setTags(s gql.FullSceneParts, videoData *VideoData) {
	for _, tag := range s.Tags {
		t := Tag{
			Name:  fmt.Sprintf("#:%s", tag.Name),
			Track: util.Ptr(0),
		}
		videoData.Tags = append(videoData.Tags, t)
	}
}

func set3DFormat(s gql.FullSceneParts, videoData *VideoData) {
	for _, tag := range s.Tags {
		switch tag.Name {
		case "DOME":
			videoData.Projection = "equirectangular"
			videoData.Stereo = "sbs"
			continue
		case "SPHERE":
			videoData.Projection = "equirectangular360"
			videoData.Stereo = "sbs"
			continue
		case "FISHEYE":
			videoData.Projection = "fisheye"
			videoData.Stereo = "sbs"
			continue
		case "MKX200":
			videoData.Projection = "fisheye"
			videoData.Stereo = "sbs"
			videoData.Lens = "MKX200"
			videoData.Fov = 200.0
			continue
		case "RF52":
			videoData.Projection = "fisheye"
			videoData.Stereo = "sbs"
			videoData.Fov = 190.0
			continue
		case "CUBEMAP":
			videoData.Projection = "cubemap"
			videoData.Stereo = "sbs"
		case "EAC":
			videoData.Projection = "equiangularCubemap"
			videoData.Stereo = "sbs"
		case "SBS":
			videoData.Stereo = "sbs"
			continue
		case "TB":
			videoData.Stereo = "tb"
			continue
		}
	}
}

func setStreamSources(s gql.FullSceneParts, videoData *VideoData) {
	for _, stream := range stash.GetStreams(s, true) {
		e := Media{
			Name: stream.Name,
		}
		for _, source := range stream.Sources {
			vs := Source{
				Resolution: source.Resolution,
				Url:        source.Url,
			}
			e.Sources = append(e.Sources, vs)
		}
		videoData.Media = append(videoData.Media, e)
	}
}
