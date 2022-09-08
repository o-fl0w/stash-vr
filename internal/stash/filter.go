package stash

import (
	"encoding/json"
	"fmt"
	"stash-vr/internal/stash/gql"
	"strconv"
)

type JsonFilter struct {
	SortBy  string   `json:"sortby"`
	SortDir string   `json:"sortdir,omitempty"`
	C       []string `json:"c"`
}

type JsonCriterion struct {
	Modifier string      `json:"modifier"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
}

type Filter struct {
	SortBy      string
	SortDir     gql.SortDirectionEnum
	SceneFilter gql.SceneFilterType
}

func ParseJsonEncodedFilter(raw string) (Filter, error) {
	var filter JsonFilter

	err := json.Unmarshal([]byte(raw), &filter)
	if err != nil {
		return Filter{}, fmt.Errorf("unmarshal: '%w'", err)
	}
	sf, err := getSceneFilter(filter.C)
	if err != nil {
		return Filter{}, fmt.Errorf("getSceneFilter: %w", err)
	}

	sortDir := gql.SortDirectionEnumAsc
	if filter.SortDir == "desc" {
		sortDir = gql.SortDirectionEnumDesc
	}
	return Filter{SortBy: filter.SortBy, SortDir: sortDir, SceneFilter: sf}, nil
}

func parseJsonCriterion(raw string) (JsonCriterion, error) {
	var c JsonCriterion

	err := json.Unmarshal([]byte(raw), &c)
	if err != nil {
		return JsonCriterion{}, fmt.Errorf("unmarshal: '%w'", err)
	}
	return c, nil
}

func getSceneFilter(jsonCriteria []string) (gql.SceneFilterType, error) {
	sf := gql.SceneFilterType{}
	for _, jsonCriterion := range jsonCriteria {
		c, err := parseJsonCriterion(jsonCriterion)
		if err != nil {
			return gql.SceneFilterType{}, fmt.Errorf("parseJsonCriterion: %w", err)
		}
		setSceneFilterCriterion(c, &sf)
	}
	return sf, nil
}

func setSceneFilterCriterion(criterion JsonCriterion, sceneFilter *gql.SceneFilterType) {
	switch criterion.Type {
	//HierarchicalMultiCriterionInput
	case "tags":
		c := parseHierarchicalMultiCriterionInput(criterion)
		sceneFilter.Tags = &c
	case "studios":
		c := parseHierarchicalMultiCriterionInput(criterion)
		sceneFilter.Studios = &c
	case "performerTags":
		c := parseHierarchicalMultiCriterionInput(criterion)
		sceneFilter.Performer_tags = &c
	//StringCriterionInput
	case "title":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Title = &c
	case "details":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Details = &c
	case "oshash":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Oshash = &c
	case "sceneChecksum":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Checksum = &c
	case "phash":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Phash = &c
	case "path":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Path = &c
	case "stash_id":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Stash_id = &c
	case "url":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Url = &c
	case "captions":
		c := parseStringCriterionInput(criterion)
		sceneFilter.Captions = &c
	//IntCriterionInput
	case "rating":
		c := parseIntCriterionInput(criterion)
		sceneFilter.Rating = &c
	case "o_counter":
		c := parseIntCriterionInput(criterion)
		sceneFilter.O_counter = &c
	case "duration":
		c := parseIntCriterionInput(criterion)
		sceneFilter.Duration = &c
	case "tag_count":
		c := parseIntCriterionInput(criterion)
		sceneFilter.Tag_count = &c
	case "performer_age":
		c := parseIntCriterionInput(criterion)
		sceneFilter.Performer_age = &c
	case "performer_count":
		c := parseIntCriterionInput(criterion)
		sceneFilter.Performer_count = &c
	case "interactive_speed":
		c := parseIntCriterionInput(criterion)
		sceneFilter.Interactive_speed = &c
	//bool
	case "organized":
		c := parseBool(criterion)
		sceneFilter.Organized = c
	case "performer_favorite":
		c := parseBool(criterion)
		sceneFilter.Performer_favorite = c
	case "interactive":
		c := parseBool(criterion)
		sceneFilter.Interactive = c
	//PHashDuplicationCriterionInput
	case "duplicated":
		c := parsePHashDuplicationCriterionInput(criterion)
		sceneFilter.Duplicated = &c
	//ResolutionCriterionInput
	case "resolution":
		c := parseResolutionCriterionInput(criterion)
		sceneFilter.Resolution = &c
		//string
	case "hasMarkers":
		c := parseString(criterion)
		sceneFilter.Has_markers = c
	case "sceneIsMissing":
		c := parseString(criterion)
		sceneFilter.Is_missing = c
	//MultiCriterionInput
	case "movies":
		c := parseMultiCriterionInput(criterion)
		sceneFilter.Movies = &c
	case "performers":
		c := parseMultiCriterionInput(criterion)
		sceneFilter.Performers = &c
	}
}

func parseHierarchicalMultiCriterionInput(c JsonCriterion) gql.HierarchicalMultiCriterionInput {
	items := c.Value.(map[string]interface{})["items"].([]interface{})
	var ids []string
	for _, item := range items {
		id := item.(map[string]interface{})["id"].(string)
		ids = append(ids, id)
	}

	return gql.HierarchicalMultiCriterionInput{
		Value:    ids,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func parseStringCriterionInput(c JsonCriterion) gql.StringCriterionInput {
	s := c.Value.(string)
	return gql.StringCriterionInput{
		Value:    s,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func parseIntCriterionInput(c JsonCriterion) gql.IntCriterionInput {
	v := c.Value.(map[string]interface{})["value"].(float64)
	_v2 := c.Value.(map[string]interface{})["value2"]
	var v2 float64
	if _v2 != nil {
		v2 = _v2.(float64)
	}
	return gql.IntCriterionInput{
		Value:    int(v),
		Value2:   int(v2),
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func parseBool(c JsonCriterion) bool {
	b, _ := strconv.ParseBool(c.Value.(string))
	return b
}

func parsePHashDuplicationCriterionInput(c JsonCriterion) gql.PHashDuplicationCriterionInput {
	b, _ := strconv.ParseBool(c.Value.(string))
	return gql.PHashDuplicationCriterionInput{
		Duplicated: b,
	}
}

func parseResolutionCriterionInput(c JsonCriterion) gql.ResolutionCriterionInput {
	s := c.Value.(string)
	var rs gql.ResolutionEnum

	switch s {
	case "144p":
		rs = gql.ResolutionEnumVeryLow
	case "240p":
		rs = gql.ResolutionEnumLow
	case "360p":
		rs = gql.ResolutionEnumR360p
	case "480p":
		rs = gql.ResolutionEnumStandard
	case "540p":
		rs = gql.ResolutionEnumWebHd
	case "720p":
		rs = gql.ResolutionEnumStandardHd
	case "1080p":
		rs = gql.ResolutionEnumFullHd
	case "1440p":
		rs = gql.ResolutionEnumQuadHd
	case "1920p":
		rs = gql.ResolutionEnumVrHd
	case "4k":
		rs = gql.ResolutionEnumFourK
	case "5k":
		rs = gql.ResolutionEnumFiveK
	case "6k":
		rs = gql.ResolutionEnumSixK
	case "8k":
		rs = gql.ResolutionEnumEightK
	}

	return gql.ResolutionCriterionInput{
		Value:    rs,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func parseString(c JsonCriterion) string {
	s := c.Value.(string)
	return s
}

func parseMultiCriterionInput(c JsonCriterion) gql.MultiCriterionInput {
	cs := c.Value.([]interface{})
	var ss []string
	for _, c := range cs {
		ss = append(ss, c.(string))
	}
	return gql.MultiCriterionInput{
		Value:    ss,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}
