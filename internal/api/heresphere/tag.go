package heresphere

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/util"
	"strconv"
	"strings"
)

type tagDto struct {
	Name   string   `json:"name"`
	Start  float64  `json:"start,omitempty"`
	End    *float64 `json:"end,omitempty"`
	Track  *int     `json:"track,omitempty"`
	Rating *float32 `json:"rating,omitempty"`
	value  string
}

const seperator = ":"

var summaryStripper, _ = regexp.Compile("[^a-zA-Z0-9_]+")

func setTrack(tracks []tagDto, track int) {
	for i := range tracks {
		tracks[i].Track = &track
	}
}

func addTrack(target *[]tagDto, tags []tagDto, track int) int {
	if len(tags) == 0 {
		return track
	}
	setTrack(tags, track)
	*target = append(*target, tags...)
	return track + 1
}

func addFullTrack(target *[]tagDto, tags []tagDto, track int, totalDuration float64) int {
	if len(tags) == 0 {
		return track
	}
	equallyDivideTagDurations(totalDuration, tags)
	setTrack(tags, track)
	*target = append(*target, tags...)
	return track + 1
}

func addMultiTracks(target *[]tagDto, tags []tagDto, track int) int {
	if len(tags) == 0 {
		return track
	}
	for _, t := range tags {
		i := track
		t.Track = &i
		*target = append(*target, t)
		track++
	}
	return track
}

func getTags(vd *library.VideoData) []tagDto {
	duration := vd.SceneParts.Files[0].Duration * 1000

	var tags []tagDto

	trackIndex := addTrack(&tags, getMarkers(vd), 0)

	summary := getSummary(vd)
	if summary != "" {
		trackIndex = addFullTrack(&tags, []tagDto{{Name: "?:" + summary}}, trackIndex, duration)
	}

	trackIndex = addFullTrack(&tags, getFields(vd), trackIndex, duration)

	trackIndex = addMultiTracks(&tags, getStashTags(vd), trackIndex)
	trackIndex = addMultiTracks(&tags, getStudio(vd), trackIndex)
	trackIndex = addMultiTracks(&tags, getPerformers(vd), trackIndex)
	trackIndex = addMultiTracks(&tags, getGroups(vd), trackIndex)

	return tags
}

func getSummary(vd *library.VideoData) string {
	if len(vd.SceneParts.Tags) == 0 {
		return ""
	}

	//tags
	m := make(map[string]string)
	for _, t := range vd.SceneParts.Tags {
		sn := t.Name
		if t.Sort_name != nil {
			if *t.Sort_name == config.Application().ExcludeSortName {
				continue
			}
			sn = *t.Sort_name
		}
		m[t.Name] = sn
	}

	type item struct {
		key     string
		sortKey string
	}

	items := make([]item, 0, len(m))
	for k, v := range m {
		items = append(items, item{key: k, sortKey: v})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].sortKey == items[j].sortKey {
			return items[i].key < items[j].key
		}
		return items[i].sortKey < items[j].sortKey
	})

	seen := make(map[string]struct{})
	keys := make([]string, 0, len(items))
	for _, it := range items {
		name := summaryStripper.ReplaceAllString(strings.ReplaceAll(it.key, " ", "_"), "")
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		keys = append(keys, it.key)
	}

	summary := strings.Join(keys, " | ")
	return summary

}

func getStashTags(vd *library.VideoData) []tagDto {
	type item struct {
		sortName string
		dto      tagDto
	}
	items := make([]item, 0, len(vd.SceneParts.Tags))
	//tags := make([]tagDto, 0, len(vd.SceneParts.Tags))
	isExcluded := func(sortName *string) bool {
		return sortName != nil && *sortName == config.Application().ExcludeSortName
	}

	for _, t := range vd.SceneParts.Tags {
		if isExcluded(t.Sort_name) {
			continue
		}
		dto := tagDto{
			Name: fmt.Sprintf("%s%s%s", internal.LegendTag, seperator, t.Name), value: t.Name,
		}
		items = append(items, item{sortName: util.FirstNonEmpty(t.Sort_name, &t.Name), dto: dto})

		for _, p := range t.Parents {
			if isExcluded(p.Sort_name) {
				continue
			}
			pDto := tagDto{
				Name: fmt.Sprintf("%s%s%s%s", internal.LegendTag, p.Name, seperator, t.Name), value: t.Name,
			}
			items = append(items, item{sortName: util.FirstNonEmpty(p.Sort_name, &p.Name), dto: pDto})
		}
	}

	slices.SortFunc(items, func(a item, b item) int {
		return strings.Compare(a.sortName, b.sortName)
	})

	tags := make([]tagDto, 0, len(items))
	for _, it := range items {
		tags = append(tags, it.dto)
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

func getStudio(vd *library.VideoData) []tagDto {
	if vd.SceneParts.Studio == nil {
		return nil
	}
	return []tagDto{{Name: fmt.Sprintf("%s%s%s", internal.LegendSceneStudio, seperator, vd.SceneParts.Studio.Name), value: vd.SceneParts.Studio.Name}}
}

func getFields(vd *library.VideoData) []tagDto {
	tags := make([]tagDto, 0)

	playCount := 0
	if vd.SceneParts.Play_count != nil {
		playCount = *vd.SceneParts.Play_count
	}
	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%d", internal.LegendMetaPlayCount, seperator, playCount)})

	oCount := 0
	if vd.SceneParts.O_counter != nil {
		oCount = *vd.SceneParts.O_counter
	}
	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%d", internal.LegendMetaOCount, seperator, oCount)})

	resolution, tier := nearestResolution(vd.SceneParts.Files[0].Height)
	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%dp", internal.LegendMetaResolution, seperator, resolution)})
	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%s", internal.LegendMetaResolution, seperator, tier)})

	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%v", internal.LegendMetaOrganized, seperator, vd.SceneParts.Organized)})

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
	n := len(tags)
	switch n {
	case 0:
	case 1:
		tags[0].Start = 1
		tags[0].End = util.Ptr(totalDuration - 1)
	default:
		durationPerItem := (totalDuration - 1) / float64(n) //-1 because HS doesn't display single full-length tags
		for i := range tags {
			tags[i].Start = 1 + float64(i)*durationPerItem
			tags[i].End = util.Ptr(float64(i+1) * durationPerItem)
		}
	}
}
