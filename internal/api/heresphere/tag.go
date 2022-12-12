package heresphere

import (
	"fmt"
	"sort"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

type tag struct {
	Name   string  `json:"name"`
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	Track  *int    `json:"track,omitempty"`
	Rating float32 `json:"rating"`
}

const seperator = ":"
const isGlanceMarkersEnabled = true

func getTags(s gql.SceneScanParts) []tag {
	var tagTracks [][]tag

	markers := getMarkers(s)
	performers := getPerformers(s)
	fields := getFields(s)

	studio := getStudio(s)
	stashTags := getStashTags(s)
	movies := getMovies(s)

	meta := make([]tag, 0, len(studio)+len(stashTags)+len(movies))
	meta = append(meta, studio...)
	meta = append(meta, stashTags...)
	meta = append(meta, movies...)

	if len(studio) == 0 {
		fields = append(fields, tag{Name: internal.LegendStudio.Full + seperator})
	}
	if len(stashTags) == 0 {
		fields = append(fields, tag{Name: internal.LegendTag.Short + seperator})
	}
	if len(movies) == 0 {
		fields = append(fields, tag{Name: internal.LegendMovie.Full + seperator})
	}

	fillTagDurations(markers)
	duration := s.Files[0].Duration * 1000
	equallyDivideTagDurations(duration, performers)
	equallyDivideTagDurations(duration, fields)
	equallyDivideTagDurations(duration, meta)

	if isGlanceMarkersEnabled {
		tagTracks = append(tagTracks, markers)
		tagTracks = append(tagTracks, meta)
	} else {
		tagTracks = append(tagTracks, meta)
		tagTracks = append(tagTracks, markers)
	}
	tagTracks = append(tagTracks, performers)
	tagTracks = append(tagTracks, fields)

	track := 0
	tags := make([]tag, 0, len(tagTracks))
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

func getPerformers(s gql.SceneScanParts) []tag {
	tags := make([]tag, len(s.Performers))
	for i, p := range s.Performers {
		tags[i] = tag{
			Name:   internal.LegendPerformer.Full + seperator + p.Name,
			Rating: float32(p.Rating100) / 20.0,
		}
	}
	return tags
}

func getMovies(s gql.SceneScanParts) []tag {
	if s.Movies == nil {
		return nil
	}
	tags := make([]tag, len(s.Movies))
	for i, m := range s.Movies {
		tags[i] = tag{
			Name: internal.LegendMovie.Full + seperator + m.Movie.Name,
		}
	}
	return tags
}

func getStudio(s gql.SceneScanParts) []tag {
	if s.Studio == nil {
		return nil
	}
	return []tag{{
		Name:   internal.LegendStudio.Full + seperator + s.Studio.Name,
		Rating: float32(s.Studio.Rating100) / 20.0,
	}}
}

func getFields(s gql.SceneScanParts) []tag {
	tags := []tag{
		{Name: fmt.Sprintf("%s:%d", internal.LegendOCount.Short, s.O_counter)},
		{Name: fmt.Sprintf("%s:%v", internal.LegendOrganized.Short, s.Organized)}}

	return tags
}

func getStashTags(s gql.SceneScanParts) []tag {
	tags := make([]tag, 0, len(s.Tags))
	for _, t := range s.Tags {
		if t.Name == config.Get().FavoriteTag {
			continue
		}
		t := tag{
			Name: internal.LegendTag.Short + seperator + t.Name,
		}
		tags = append(tags, t)
	}
	return tags
}

func getMarkers(s gql.SceneScanParts) []tag {
	tags := make([]tag, len(s.Scene_markers))
	for i, sm := range s.Scene_markers {
		tagName := sm.Primary_tag.Name
		if sm.Title != "" {
			tagName += seperator + sm.Title
		}
		t := tag{
			Name:  tagName,
			Start: sm.Seconds * 1000,
		}
		tags[i] = t
	}
	return tags
}

func equallyDivideTagDurations(totalDuration float64, tags []tag) {
	durationPerItem := totalDuration / float64(len(tags))
	for i := range tags {
		tags[i].Start = float64(i) * durationPerItem
		tags[i].End = float64(i+1) * durationPerItem
	}
}

func fillTagDurations(tags []tag) {
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
