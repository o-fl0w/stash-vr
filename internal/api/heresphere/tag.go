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

func setTrack(tag *tagDto, track int) {
	tag.Track = &track
}

func setTracks(tags []tagDto, track int) {
	for i := range tags {
		setTrack(&tags[i], track)
	}
}

func addTrack(target *[]tagDto, tags []tagDto, track int) int {
	if len(tags) == 0 {
		return track
	}
	setTracks(tags, track)
	*target = append(*target, tags...)
	return track + 1
}

func addSplitTrack(target *[]tagDto, tags []tagDto, track int, totalDuration float64) int {
	if len(tags) == 0 {
		return track
	}
	equallyDivideTagDurations(totalDuration, tags)
	setTracks(tags, track)
	*target = append(*target, tags...)
	return track + 1
}

func addMultiTracks(target *[]tagDto, tags []tagDto, startTrack int) int {
	tagCount := len(tags)
	if tagCount == 0 {
		return startTrack
	}
	for i, t := range tags {
		setTrack(&t, startTrack+i)
		*target = append(*target, t)
	}
	return startTrack + tagCount
}

func getTags(vd *library.VideoData) []tagDto {
	duration := vd.SceneParts.Files[0].Duration * 1000

	var tags []tagDto

	trackIndex := addTrack(&tags, getMarkers(vd), 0)

	summary := getSummary(vd)
	if summary != "" {
		trackIndex = addSplitTrack(&tags, []tagDto{{Name: internal.LegendSummary + seperator + summary}}, trackIndex, duration)
	}

	trackIndex = addSplitTrack(&tags, getFields(vd), trackIndex, duration)

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
		if t.Sort_name == config.Application().ExcludeSortName {
			continue
		}
		m[t.Name] = util.FirstNonEmpty(&t.Sort_name, &t.Name)
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
		keys = append(keys, name)
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

	for _, t := range vd.SceneParts.Tags {
		if t.Sort_name == config.Application().ExcludeSortName {
			continue
		}
		dto := tagDto{
			Name: fmt.Sprintf("%s%s%s", internal.LegendTag, seperator, t.Name), value: t.Name,
		}
		tSortName := util.FirstNonEmpty(&t.Sort_name, &t.Name)
		items = append(items, item{sortName: tSortName, dto: dto})

		for _, p := range t.Parents {
			if p.Sort_name == config.Application().ExcludeSortName {
				continue
			}
			pDto := tagDto{
				Name: fmt.Sprintf("%s%s%s%s", internal.LegendTag, p.Name, seperator, t.Name), value: t.Name,
			}
			items = append(items, item{sortName: util.FirstNonEmpty(&p.Sort_name, &p.Name), dto: pDto})
		}
	}
	parentItems := make([]item, 0)
	childItems := make([]item, 0)
	for _, it := range items {
		if strings.HasSuffix(it.sortName, "#") {
			parentItems = append(parentItems, it)
		} else {
			childItems = append(childItems, it)
		}
	}

	slices.SortFunc(parentItems, func(a item, b item) int {
		if a.sortName == b.sortName {
			return strings.Compare(a.dto.Name, b.dto.Name)
		}
		return strings.Compare(a.sortName, b.sortName)
	})
	slices.SortFunc(childItems, func(a item, b item) int {
		if a.sortName == b.sortName {
			return strings.Compare(a.dto.Name, b.dto.Name)
		}
		return strings.Compare(a.sortName, b.sortName)
	})

	tags := make([]tagDto, 0, len(items))
	for _, it := range parentItems {
		tags = append(tags, it.dto)
	}
	for _, it := range childItems {
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

	rating := "?"
	if vd.SceneParts.Rating100 != nil {
		rating = strconv.Itoa(*vd.SceneParts.Rating100)
	}
	tags = append(tags, tagDto{Name: fmt.Sprintf("%s%s%s", internal.LegendMetaRating, seperator, rating)})

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
		tags[0].Start = 0.1
	default:
		durationPerItem := (totalDuration - 0.1) / float64(n) //-1 because HS doesn't display single full-length tags
		for i := range tags {
			start := 0.1 + float64(i)*durationPerItem
			end := start + durationPerItem
			tags[i].Start = start
			tags[i].End = &end
		}
	}
}
