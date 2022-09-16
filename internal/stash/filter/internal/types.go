package internal

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

func ParseJsonCriterion(raw string) (JsonCriterion, error) {
	var c JsonCriterion

	err := json.Unmarshal([]byte(raw), &c)
	if err != nil {
		return JsonCriterion{}, fmt.Errorf("unmarshal json criterion '%s': %w", raw, err)
	}
	return c, nil
}

func (c JsonCriterion) AsHierarchicalMultiCriterionInput() *gql.HierarchicalMultiCriterionInput {
	items := c.Value.(map[string]interface{})["items"].([]interface{})
	var ids []string
	for _, item := range items {
		id := item.(map[string]interface{})["id"].(string)
		ids = append(ids, id)
	}

	return &gql.HierarchicalMultiCriterionInput{
		Value:    ids,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func (c JsonCriterion) AsStringCriterionInput() *gql.StringCriterionInput {
	s := c.Value.(string)
	return &gql.StringCriterionInput{
		Value:    s,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func (c JsonCriterion) AsIntCriterionInput() *gql.IntCriterionInput {
	v := c.Value.(map[string]interface{})["value"].(float64)
	_v2 := c.Value.(map[string]interface{})["value2"]
	var v2 float64
	if _v2 != nil {
		v2 = _v2.(float64)
	}
	return &gql.IntCriterionInput{
		Value:    int(v),
		Value2:   int(v2),
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func (c JsonCriterion) AsBool() bool {
	b, _ := strconv.ParseBool(c.Value.(string))
	return b
}

func (c JsonCriterion) AsPHashDuplicationCriterionInput() *gql.PHashDuplicationCriterionInput {
	b, _ := strconv.ParseBool(c.Value.(string))
	return &gql.PHashDuplicationCriterionInput{
		Duplicated: b,
	}
}

func (c JsonCriterion) AsResolutionCriterionInput() *gql.ResolutionCriterionInput {
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

	return &gql.ResolutionCriterionInput{
		Value:    rs,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}

func (c JsonCriterion) AsString() string {
	s := c.Value.(string)
	return s
}

func (c JsonCriterion) AsMultiCriterionInput() *gql.MultiCriterionInput {
	cs := c.Value.([]interface{})
	var ss []string
	for _, c := range cs {
		ss = append(ss, c.(string))
	}
	return &gql.MultiCriterionInput{
		Value:    ss,
		Modifier: gql.CriterionModifier(c.Modifier),
	}
}
