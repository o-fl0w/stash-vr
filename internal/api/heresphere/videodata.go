package heresphere

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"sort"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/heresphere/proto"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"strings"
)

func buildVideoData(ctx context.Context, client graphql.Client, sceneId string) (proto.VideoData, error) {
	findSceneResponse, err := gql.FindScene(ctx, client, sceneId)
	if err != nil {
		return proto.VideoData{}, fmt.Errorf("FindScene: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return proto.VideoData{}, fmt.Errorf("FindScene: not found")
	}
	s := findSceneResponse.FindScene.FullSceneParts

	videoData := proto.VideoData{
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

func setTags(s gql.FullSceneParts, videoData *proto.VideoData) {
	tags := getTags(s)
	videoData.Tags = tags
}

func getTags(s gql.FullSceneParts) []proto.Tag {
	var tagTracks [][]proto.Tag

	markers := getMarkers(s)
	performers := getPerformers(s)
	fields := getFields(s)

	var meta []proto.Tag
	studio := getStudio(s)
	stashTags := getStashTags(s)
	movies := getMovies(s)

	meta = append(meta, studio...)
	meta = append(meta, stashTags...)
	meta = append(meta, movies...)

	if len(studio) == 0 {
		fields = append(fields, proto.Tag{Name: fmt.Sprintf("%s:", common.LegendStudio.Full)})
	}
	if len(stashTags) == 0 {
		fields = append(fields, proto.Tag{Name: fmt.Sprintf("%s:", common.LegendTag.Short)})
	}
	if len(movies) == 0 {
		fields = append(fields, proto.Tag{Name: fmt.Sprintf("%s:", common.LegendMovie.Full)})
	}

	fillTagDurations(markers)
	duration := s.File.Duration * 1000
	equallyDivideTagDurations(duration, performers)
	equallyDivideTagDurations(duration, fields)
	equallyDivideTagDurations(duration, meta)

	if config.Get().IsGlanceMarkersEnabled {
		tagTracks = append(tagTracks, markers)
		tagTracks = append(tagTracks, meta)
	} else {
		tagTracks = append(tagTracks, meta)
		tagTracks = append(tagTracks, markers)
	}
	tagTracks = append(tagTracks, performers)
	tagTracks = append(tagTracks, fields)

	var tags []proto.Tag
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

func getPerformers(s gql.FullSceneParts) []proto.Tag {
	tags := make([]proto.Tag, len(s.Performers))
	for i, p := range s.Performers {
		tags[i] = proto.Tag{
			Name:   fmt.Sprintf("%s:%s", common.LegendPerformer.Full, p.Name),
			Rating: float32(p.Rating),
		}
	}
	return tags
}

func getMovies(s gql.FullSceneParts) []proto.Tag {
	if s.Movies == nil {
		return nil
	}
	tags := make([]proto.Tag, len(s.Movies))
	for i, m := range s.Movies {
		tags[i] = proto.Tag{
			Name: fmt.Sprintf("%s:%s", common.LegendMovie.Full, m.Movie.Name),
		}
	}
	return tags
}

func getStudio(s gql.FullSceneParts) []proto.Tag {
	if s.Studio == nil {
		return nil
	}
	return []proto.Tag{{
		Name:   fmt.Sprintf("%s:%s", common.LegendStudio.Full, s.Studio.Name),
		Rating: float32(s.Studio.Rating),
	}}
}

func getFields(s gql.FullSceneParts) []proto.Tag {
	var tags []proto.Tag

	tags = append(tags, proto.Tag{
		Name: fmt.Sprintf("%s:%d", common.LegendOCount.Short, s.O_counter),
	})

	tags = append(tags, proto.Tag{
		Name: fmt.Sprintf("%s:%v", common.LegendOrganized.Short, s.Organized),
	})

	return tags
}

func getStashTags(s gql.FullSceneParts) []proto.Tag {
	var tags []proto.Tag
	for _, tag := range s.Tags {
		if tag.Name == config.Get().FavoriteTag {
			continue
		}
		t := proto.Tag{
			Name: fmt.Sprintf("%s:%s", common.LegendTag.Short, tag.Name),
		}
		tags = append(tags, t)
	}
	return tags
}

func getMarkers(s gql.FullSceneParts) []proto.Tag {
	var tags []proto.Tag
	for _, sm := range s.Scene_markers {
		sb := strings.Builder{}
		sb.WriteString(sm.Primary_tag.Name)
		if sm.Title != "" {
			sb.WriteString(":")
			sb.WriteString(sm.Title)
		}
		t := proto.Tag{
			Name:  sb.String(),
			Start: int(sm.Seconds * 1000),
		}
		tags = append(tags, t)
	}
	return tags
}

func equallyDivideTagDurations(totalDuration float64, tags []proto.Tag) {
	durationPerItem := int(totalDuration / float64(len(tags)))
	for i := range tags {
		tags[i].Start = i * durationPerItem
		tags[i].End = (i + 1) * durationPerItem
	}
}

func fillTagDurations(tags []proto.Tag) {
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

func setScripts(s gql.FullSceneParts, videoData *proto.VideoData) {
	if s.ScriptParts.Interactive {
		videoData.Scripts = append(videoData.Scripts, proto.Script{
			Name: fmt.Sprintf("Script-%s", s.Title),
			Url:  s.ScriptParts.Paths.Funscript,
		})
	}
}

func set3DFormat(s gql.FullSceneParts, videoData *proto.VideoData) {
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

func setStreamSources(ctx context.Context, s gql.FullSceneParts, videoData *proto.VideoData) {
	for _, stream := range stash.GetStreams(ctx, s, true) {
		e := proto.Media{
			Name: stream.Name,
		}
		for _, source := range stream.Sources {
			vs := proto.Source{
				Resolution: source.Resolution,
				Url:        source.Url,
			}
			e.Sources = append(e.Sources, vs)
		}
		videoData.Media = append(videoData.Media, e)
	}
}

func setIsFavorite(s gql.FullSceneParts, videoData *proto.VideoData) {
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
