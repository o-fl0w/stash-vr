package videodata

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"stash-vr/internal/api/heresphere/internal/tag"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

type VideoData struct {
	Access int `json:"access"`

	Title          string    `json:"title"`
	Description    string    `json:"description"`
	ThumbnailImage string    `json:"thumbnailImage"`
	ThumbnailVideo string    `json:"thumbnailVideo"`
	DateReleased   string    `json:"dateReleased"`
	DateAdded      string    `json:"dateAdded"`
	Duration       float64   `json:"duration"`
	Rating         float32   `json:"rating"`
	Favorites      int       `json:"favorites"`
	IsFavorite     bool      `json:"isFavorite"`
	Projection     string    `json:"projection"`
	Stereo         string    `json:"stereo"`
	Fov            float32   `json:"fov"`
	Lens           string    `json:"lens"`
	Scripts        []Script  `json:"scripts"`
	Tags           []tag.Tag `json:"tags"`
	Media          []Media   `json:"media"`

	WriteFavorite bool `json:"writeFavorite"`
	WriteRating   bool `json:"writeRating"`
	WriteTags     bool `json:"writeTags"`
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

func Build(ctx context.Context, client graphql.Client, sceneId string) (VideoData, error) {
	findSceneResponse, err := gql.FindSceneFull(ctx, client, sceneId)
	if err != nil {
		return VideoData{}, fmt.Errorf("FindSceneFull: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return VideoData{}, fmt.Errorf("FindSceneFull: not found")
	}
	s := findSceneResponse.FindScene.SceneFullParts

	videoData := VideoData{
		Access:         1,
		Title:          s.Title,
		Description:    s.Details,
		ThumbnailImage: stash.ApiKeyed(s.Paths.Screenshot),
		ThumbnailVideo: stash.ApiKeyed(s.Paths.Preview),
		DateReleased:   s.Date,
		DateAdded:      s.Created_at.Format("2006-01-02"),
		Duration:       s.SceneDetailsParts.File.Duration * 1000,
		Rating:         float32(s.Rating),
		Favorites:      s.O_counter,
		WriteFavorite:  true,
		WriteRating:    true,
		WriteTags:      true,
	}

	setIsFavorite(s, &videoData)

	setStreamSources(ctx, s, &videoData)
	set3DFormat(s, &videoData)

	setTags(s, &videoData)

	setScripts(s, &videoData)

	return videoData, nil
}

func setTags(s gql.SceneFullParts, videoData *VideoData) {
	tags := tag.GetTags(s.SceneDetailsParts)
	videoData.Tags = tags
}

func setScripts(s gql.SceneFullParts, videoData *VideoData) {
	if s.ScriptParts.Interactive {
		videoData.Scripts = append(videoData.Scripts, Script{
			Name: fmt.Sprintf("Script-%s", s.Title),
			Url:  s.ScriptParts.Paths.Funscript,
		})
	}
}

func set3DFormat(s gql.SceneFullParts, videoData *VideoData) {
	for _, t := range s.Tags {
		switch t.Name {
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

func setStreamSources(ctx context.Context, s gql.SceneFullParts, videoData *VideoData) {
	for _, stream := range stash.GetStreams(ctx, s, true) {
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

func setIsFavorite(s gql.SceneFullParts, videoData *VideoData) {
	videoData.IsFavorite = ContainsFavoriteTag(s.TagPartsArray)
}

func ContainsFavoriteTag(ts gql.TagPartsArray) bool {
	for _, t := range ts.Tags {
		if t.Name == config.Get().FavoriteTag {
			return true
		}
	}
	return false
}
