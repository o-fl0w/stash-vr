package heresphere

import (
	"fmt"
	"slices"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/util"
	"strconv"
)

type tagDto struct {
	Name   string   `json:"name"`
	Start  float64  `json:"start,omitempty"`
	End    *float64 `json:"end,omitempty"`
	Track  *int     `json:"track,omitempty"`
	Rating *float32 `json:"rating,omitempty"`
}

const seperator = ":"

func getTags(vd *library.VideoData) []tagDto {
	var tracks = make([][]tagDto, 0, 6)

	duration := vd.SceneParts.Files[0].Duration * 1000

	if stashTags := getStashTags(vd); len(stashTags) > 0 {
		equallyDivideTagDurations(duration, stashTags)
		tracks = append(tracks, stashTags)
	}

	if performers := getPerformers(vd); len(performers) > 0 {
		equallyDivideTagDurations(duration, performers)
		tracks = append(tracks, performers)
	}

	sceneData := getGroups(vd)
	if studio := getStudio(vd); studio != nil {
		sceneData = slices.Insert(sceneData, 0, *studio)
	}
	if len(sceneData) > 0 {
		equallyDivideTagDurations(duration, sceneData)
		tracks = append(tracks, sceneData)
	}

	if metaFields := getFields(vd); len(metaFields) > 0 {
		equallyDivideTagDurations(duration, metaFields)
		tracks = append(tracks, metaFields)
	}

	markers := getMarkers(vd)
	for _, marker := range markers {
		tracks = append(tracks, []tagDto{marker})
	}

	tags := make([]tagDto, 0, len(tracks))
	for i := range tracks {
		for j := range tracks[i] {
			track := i
			tag := tracks[i][j]
			tag.Track = &track
			tags = append(tags, tag)
		}
	}

	return tags
}

func getPerformers(vd *library.VideoData) []tagDto {
	tags := make([]tagDto, len(vd.SceneParts.Performers))
	for i, p := range vd.SceneParts.Performers {
		tags[i] = tagDto{
			Name: fmt.Sprintf("%s%s%s", internal.LegendPerformer, seperator, p.Name),
		}
	}
	return tags
}

func getGroups(vd *library.VideoData) []tagDto {
	if vd.SceneParts.Groups == nil {
		return nil
	}
	tags := make([]tagDto, len(vd.SceneParts.Groups))
	for i, m := range vd.SceneParts.Groups {
		tags[i] = tagDto{
			Name: fmt.Sprintf("%s%s%s", internal.LegendSceneGroup, seperator, m.Group.Name),
		}
	}
	return tags
}

func getStudio(vd *library.VideoData) *tagDto {
	if vd.SceneParts.Studio == nil {
		return nil
	}
	return &tagDto{Name: fmt.Sprintf("%s%s%s", internal.LegendSceneStudio, seperator, vd.SceneParts.Studio.Name)}
}

func getFields(vd *library.VideoData) []tagDto {
	tags := make([]tagDto, 0)
	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%v", internal.LegendMetaOrganized, seperator, vd.SceneParts.Organized)})

	if vd.SceneParts.Play_count != nil {
		tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%d", internal.LegendMetaPlayCount, seperator, *vd.SceneParts.Play_count)})
	}
	if vd.SceneParts.O_counter != nil {
		tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%d", internal.LegendMetaOCount, seperator, *vd.SceneParts.O_counter)})
	}

	return tags
}

func getStashTags(vd *library.VideoData) []tagDto {
	tags := make([]tagDto, 0, len(vd.SceneParts.Tags))
	for _, t := range vd.SceneParts.Tags {
		if t.Name == config.Get().FavoriteTag {
			continue
		}
		t := tagDto{
			Name: fmt.Sprintf("%s%s%s", internal.LegendTag, seperator, t.Name),
		}
		tags = append(tags, t)
	}
	return tags
}

func getMarkers(vd *library.VideoData) []tagDto {
	tags := make([]tagDto, len(vd.SceneParts.Scene_markers))
	for i, sm := range vd.SceneParts.Scene_markers {
		tagName := sm.Primary_tag.Name
		if sm.Title != "" {
			tagName += seperator + sm.Title
		}
		var endSeconds *float64
		if sm.End_seconds != nil {
			endSeconds = util.Ptr(*sm.End_seconds * 1000)
		}

		markerId, _ := strconv.ParseFloat(sm.Id, 32)

		t := tagDto{
			Name:   tagName,
			Start:  sm.Seconds * 1000,
			End:    endSeconds,
			Rating: util.Ptr(float32(markerId)),
		}
		tags[i] = t
	}
	return tags
}

func equallyDivideTagDurations(totalDuration float64, tags []tagDto) {
	durationPerItem := (totalDuration - 1) / float64(len(tags)) //-1 because HS doesn't display single full-length tags
	for i := range tags {
		tags[i].Start = 1 + float64(i)*durationPerItem
		tags[i].End = util.Ptr(float64(i+1) * durationPerItem)
	}
}
