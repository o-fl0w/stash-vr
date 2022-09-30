package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"sort"
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
	legendMovie     = NewLegend("*", "Movie")
	legendOCount    = NewLegend("O", "O-Count")
	legendOrganized = NewLegend("Org", "Organized")
)

type VideoData struct {
	Access int `json:"access"`

	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ThumbnailImage string   `json:"thumbnailImage"`
	ThumbnailVideo string   `json:"thumbnailVideo"`
	DateReleased   string   `json:"dateReleased"`
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
		return VideoData{}, fmt.Errorf("FindScene: not found")
	}
	s := findSceneResponse.FindScene.FullSceneParts

	videoData := VideoData{
		Access:         1,
		Title:          s.Title,
		Description:    s.Details,
		ThumbnailImage: stash.ApiKeyed(s.Paths.Screenshot),
		ThumbnailVideo: stash.ApiKeyed(s.Paths.Preview),
		DateReleased:   s.Date,
		DateAdded:      s.Created_at.Format("2006-01-02"),
		Duration:       int(s.File.Duration) * 1000,
		Rating:         float32(s.Rating),
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

func setTags(s gql.FullSceneParts, videoData *VideoData) {
	tags := getTags(s)
	videoData.Tags = tags
}

func getTags(s gql.FullSceneParts) []Tag {
	var tagTracks [][]Tag

	markers := getMarkers(s)
	performers := getPerformers(s)
	fields := getFields(s)

	var meta []Tag
	studio := getStudio(s)
	stashTags := getStashTags(s)
	movies := getMovies(s)

	meta = append(meta, studio...)
	meta = append(meta, stashTags...)
	meta = append(meta, movies...)

	if len(studio) == 0 {
		fields = append(fields, Tag{Name: fmt.Sprintf("%s:", legendStudio.Full)})
	}
	if len(stashTags) == 0 {
		fields = append(fields, Tag{Name: fmt.Sprintf("%s:", legendTag.Short)})
	}
	if len(movies) == 0 {
		fields = append(fields, Tag{Name: fmt.Sprintf("%s:", legendMovie.Full)})
	}

	fillTagDurations(markers)
	duration := s.File.Duration * 1000
	equallyDivideTagDurations(duration, performers)
	equallyDivideTagDurations(duration, fields)
	equallyDivideTagDurations(duration, meta)

	if config.Get().HeresphereQuickMarkers {
		tagTracks = append(tagTracks, markers)
		tagTracks = append(tagTracks, meta)
	} else {
		tagTracks = append(tagTracks, meta)
		tagTracks = append(tagTracks, markers)
	}
	tagTracks = append(tagTracks, performers)
	tagTracks = append(tagTracks, fields)

	var tags []Tag
	track := 0
	for i := range tagTracks {
		if len(tagTracks[i]) == 0 {
			continue
		}
		for j := range tagTracks[i] {
			tagTracks[i][j].Track = util.Ptr(track)
			tags = append(tags, tagTracks[i][j])
		}
		track++
	}
	return tags
}

func getPerformers(s gql.FullSceneParts) []Tag {
	tags := make([]Tag, len(s.Performers))
	for i, p := range s.Performers {
		tags[i] = Tag{
			Name:   fmt.Sprintf("%s:%s", legendPerformer.Full, p.Name),
			Rating: float32(p.Rating),
		}
	}
	return tags
}

func getMovies(s gql.FullSceneParts) []Tag {
	if s.Movies == nil {
		return nil
	}
	tags := make([]Tag, len(s.Movies))
	for i, m := range s.Movies {
		tags[i] = Tag{
			Name: fmt.Sprintf("%s:%s", legendMovie.Full, m.Movie.Name),
		}
	}
	return tags
}

func getStudio(s gql.FullSceneParts) []Tag {
	if s.Studio == nil {
		return nil
	}
	return []Tag{{
		Name:   fmt.Sprintf("%s:%s", legendStudio.Full, s.Studio.Name),
		Rating: float32(s.Studio.Rating),
	}}
}

func getFields(s gql.FullSceneParts) []Tag {
	var tags []Tag

	tags = append(tags, Tag{
		Name: fmt.Sprintf("%s:%d", legendOCount.Short, s.O_counter),
	})

	tags = append(tags, Tag{
		Name: fmt.Sprintf("%s:%v", legendOrganized.Short, s.Organized),
	})

	return tags
}

func getStashTags(s gql.FullSceneParts) []Tag {
	var tags []Tag
	for _, tag := range s.Tags {
		if tag.Name == config.Get().FavoriteTag {
			continue
		}
		t := Tag{
			Name: fmt.Sprintf("%s:%s", legendTag.Short, tag.Name),
		}
		tags = append(tags, t)
	}
	return tags
}

func getMarkers(s gql.FullSceneParts) []Tag {
	var tags []Tag
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
		}
		tags = append(tags, t)
	}
	return tags
}

func equallyDivideTagDurations(totalDuration float64, tags []Tag) {
	durationPerItem := int(totalDuration / float64(len(tags)))
	for i := range tags {
		tags[i].Start = i * durationPerItem
		tags[i].End = (i + 1) * durationPerItem
	}
}

func fillTagDurations(tags []Tag) {
	sort.Slice(tags, func(i, j int) bool { return tags[i].Start < tags[j].Start })
	for i := range tags {
		if i == len(tags)-1 {
			tags[i].End = 0
		} else if tags[i+1].Start == 0 {
			tags[i].End = tags[i].Start + 20000
		} else {
			tags[i].End = tags[i+1].Start
		}
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

func setStreamSources(ctx context.Context, s gql.FullSceneParts, videoData *VideoData) {
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

func setIsFavorite(s gql.FullSceneParts, videoData *VideoData) {
	videoData.IsFavorite = containsFavoriteTag(s.TagPartsArray)
}

func containsFavoriteTag(ts gql.TagPartsArray) bool {
	for _, t := range ts.Tags {
		if t.Name == config.Get().FavoriteTag {
			return true
		}
	}
	return false
}
