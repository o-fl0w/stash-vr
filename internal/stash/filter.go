package stash

import (
	"encoding/json"
	"fmt"
	"stash-vr/internal/logger"
	"stash-vr/internal/stash/gql"
	"strconv"
)

type Criterion struct {
	modifier string
	typ      string
	value    interface{}
}

func ParseJsonFilter(j string) (gql.SceneFilterType, error) {
	criteria, err := parseFilter(j)
	if err != nil {
		logger.Log.Error().Str("input", j).Msg("Failed to parse Json filter")
		return gql.SceneFilterType{}, fmt.Errorf("parseFilter: %w", err)
	}

	sceneFilter := gql.SceneFilterType{}
	for _, rawCriterion := range criteria {
		criterion := parseRawCriterion(rawCriterion.(string))
		parseCriterion(criterion, &sceneFilter)
	}
	return sceneFilter, nil
}

func parseFilter(f string) ([]interface{}, error) {
	var filter interface{}

	err := json.Unmarshal([]byte(f), &filter)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: '%w'", err)
	}
	return filter.(map[string]interface{})["c"].([]interface{}), nil
}

func parseRawCriterion(raw string) Criterion {
	var c interface{}
	_ = json.Unmarshal([]byte(raw), &c)

	cmap := c.(map[string]interface{})

	return Criterion{
		modifier: cmap["modifier"].(string),
		typ:      cmap["type"].(string),
		value:    cmap["value"].(interface{}),
	}
}

func parseCriterion(criterion Criterion, sceneFilter *gql.SceneFilterType) {
	switch criterion.typ {
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

func parseHierarchicalMultiCriterionInput(c Criterion) gql.HierarchicalMultiCriterionInput {
	items := c.value.(map[string]interface{})["items"].([]interface{})
	var ids []string
	for _, item := range items {
		id := item.(map[string]interface{})["id"].(string)
		ids = append(ids, id)
	}

	return gql.HierarchicalMultiCriterionInput{
		Value:    ids,
		Modifier: gql.CriterionModifier(c.modifier),
	}
}

func parseStringCriterionInput(c Criterion) gql.StringCriterionInput {
	s := c.value.(string)
	return gql.StringCriterionInput{
		Value:    s,
		Modifier: gql.CriterionModifier(c.modifier),
	}
}

func parseIntCriterionInput(c Criterion) gql.IntCriterionInput {
	v := c.value.(map[string]interface{})["value"].(float64)
	_v2 := c.value.(map[string]interface{})["value2"]
	var v2 float64
	if _v2 != nil {
		v2 = _v2.(float64)
	}
	return gql.IntCriterionInput{
		Value:    int(v),
		Value2:   int(v2),
		Modifier: gql.CriterionModifier(c.modifier),
	}
}

func parseBool(c Criterion) bool {
	b, _ := strconv.ParseBool(c.value.(string))
	return b
}

func parsePHashDuplicationCriterionInput(c Criterion) gql.PHashDuplicationCriterionInput {
	b, _ := strconv.ParseBool(c.value.(string))
	return gql.PHashDuplicationCriterionInput{
		Duplicated: b,
	}
}

func parseResolutionCriterionInput(c Criterion) gql.ResolutionCriterionInput {
	s := c.value.(string)
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
		Modifier: gql.CriterionModifier(c.modifier),
	}
}

func parseString(c Criterion) string {
	s := c.value.(string)
	return s
}

func parseMultiCriterionInput(c Criterion) gql.MultiCriterionInput {
	cs := c.value.([]interface{})
	var ss []string
	for _, c := range cs {
		ss = append(ss, c.(string))
	}
	return gql.MultiCriterionInput{
		Value:    ss,
		Modifier: gql.CriterionModifier(c.modifier),
	}
}
