package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
)

var (
	legendTag       = NewLegend("#", "Tag")
	legendStudio    = NewLegend("$", "Studio")
	legendPerformer = NewLegend("@", "Performer")
)

type VideoData struct {
	Access int `json:"access"`

	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ThumbnailImage string   `json:"thumbnailImage"`
	ThumbnailVideo string   `json:"thumbnailVideo"`
	DateAdded      string   `json:"dateAdded"`
	Duration       int      `json:"duration"`
	Rating         float32  `json:"rating"`
	IsFavorite     bool     `json:"isFavorite"`
	Projection     string   `json:"projection"`
	Stereo         string   `json:"stereo"`
	Fov            float32  `json:"fov"`
	Lens           string   `json:"lens"`
	Scripts        []Script `json:"scripts"`
	Tags           []Tag    `json:"tags"`
	Media          []Media  `json:"media"`

	WriteFavorite bool `json:"writeFavorite"`
	WriteRating   bool `json:"writeRating"`
	WriteTags     bool `json:"writeTags"`
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

func buildVideoData(ctx context.Context, client graphql.Client, sceneId string) (VideoData, error) {
	findSceneResponse, err := gql.FindScene(ctx, client, sceneId)
	if err != nil {
		return VideoData{}, fmt.Errorf("FindScene: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return VideoData{}, fmt.Errorf("find scene: video not found")
	}
	s := findSceneResponse.FindScene.FullSceneParts

	videoData := VideoData{
		Access:         1,
		Title:          s.Title,
		Description:    s.Details,
		ThumbnailImage: stash.ApiKeyed(s.Paths.Screenshot),
		ThumbnailVideo: stash.ApiKeyed(s.Paths.Preview),
		DateAdded:      s.Created_at.Format("2006-01-02"),
		Duration:       int(s.File.Duration) * 1000,
		Rating:         float32(s.Rating),
		WriteFavorite:  true,
		WriteRating:    true,
		WriteTags:      true,
	}

	setIsFavorite(s, &videoData)

	setStreamSources(s, &videoData)
	set3DFormat(s, &videoData)

	setStudioAndTags(s, &videoData)
	setPerformers(s, &videoData)
	setMarkers(s, &videoData)
	setScripts(s, &videoData)

	return videoData, nil
}

func setStudioAndTags(s gql.FullSceneParts, videoData *VideoData) {
	itemCount := 1 + len(s.Tags)
	durationPerItem := int(s.File.Duration * 1000 / float64(itemCount))

	if s.Studio != nil {
		t := Tag{
			Name:   fmt.Sprintf("%s:%s", legendStudio.Full, s.Studio.Name),
			Rating: float32(s.Studio.Rating),
			Start:  0,
			End:    durationPerItem,
			Track:  util.Ptr(0),
		}
		videoData.Tags = append(videoData.Tags, t)
	}

	for i, tag := range s.Tags {
		t := Tag{
			Name:  fmt.Sprintf("%s:%s", legendTag.Short, tag.Name),
			Start: durationPerItem + i*durationPerItem,
			End:   durationPerItem + (i+1)*durationPerItem,
			Track: util.Ptr(0),
		}
		videoData.Tags = append(videoData.Tags, t)
	}
}

func setPerformers(s gql.FullSceneParts, videoData *VideoData) {
	itemCount := len(s.Performers)
	durationPerItem := int(s.File.Duration * 1000 / float64(itemCount))
	for i, p := range s.Performers {
		t := Tag{
			Name:   fmt.Sprintf("%s:%s", legendPerformer.Full, p.Name),
			Start:  i * durationPerItem,
			End:    (i + 1) * durationPerItem,
			Track:  util.Ptr(1),
			Rating: float32(p.Rating),
		}
		videoData.Tags = append(videoData.Tags, t)
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
		t := Tag{
			Name:  sb.String(),
			Start: int(sm.Seconds * 1000),
			End:   0,
			//Track: util.Ptr(0),
		}
		videoData.Tags = append(videoData.Tags, t)
	}
}

func setScripts(s gql.FullSceneParts, videoData *VideoData) {
	if s.ScriptParts.Interactive {
		videoData.Scripts = append(videoData.Scripts, Script{
			Name: fmt.Sprintf("Script-%s", s.Title),
			Url:  s.ScriptParts.Paths.Funscript,
		})
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

func setIsFavorite(s gql.FullSceneParts, videoData *VideoData) {
	for _, t := range s.Tags {
		if t.Name == config.Get().FavoriteTag {
			videoData.IsFavorite = true
			return
		}
	}
}
