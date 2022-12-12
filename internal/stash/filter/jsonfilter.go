package filter

import (
	"encoding/json"
	"fmt"
	"stash-vr/internal/stash/gql"
	"strconv"
)

type jsonFilter struct {
	SortBy  string   `json:"sortby"`
	SortDir string   `json:"sortdir,omitempty"`
	C       []string `json:"c"`
}

type jsonCriterion struct {
	Modifier string      `json:"modifier"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
}

type errUnexpectedType struct {
	sourceType      string
	destinationType string
}

func (e errUnexpectedType) Error() string {
	return fmt.Sprintf("unexpected type %s is not assertable to %s", e.sourceType, e.destinationType)
}

func newUnexpectedTypeErr(source any, destinationType string) *errUnexpectedType {
	return &errUnexpectedType{
		sourceType:      fmt.Sprintf("%T", source),
		destinationType: destinationType,
	}
}

func parseJsonCriterion(raw string) (jsonCriterion, error) {
	var c jsonCriterion

	err := json.Unmarshal([]byte(raw), &c)
	if err != nil {
		return jsonCriterion{}, fmt.Errorf("unmarshal json criterion '%s': %w", raw, err)
	}
	return c, nil
}

func (c jsonCriterion) asHierarchicalMultiCriterionInput() (*gql.HierarchicalMultiCriterionInput, error) {
	m, ok := c.Value.(map[string]interface{})
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value, "map[string]interface{}")
	}
	items, ok := m["items"].([]interface{})
	if !ok {
		return nil, newUnexpectedTypeErr(m["items"], "[]interface{}")
	}
	ids := make([]string, len(items))
	for i, item := range items {
		mid, ok := item.(map[string]interface{})
		if !ok {
			return nil, newUnexpectedTypeErr(item, "map[string]interface{}")
		}
		id, ok := mid["id"].(string)
		if !ok {
			return nil, newUnexpectedTypeErr(mid["id"], "string")
		}
		ids[i] = id
	}

	return &gql.HierarchicalMultiCriterionInput{
		Value:    ids,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asStringCriterionInput() (*gql.StringCriterionInput, error) {
	if c.isNullable() {
		return &gql.StringCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}

	s, ok := c.Value.(string)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value, "string")
	}
	return &gql.StringCriterionInput{
		Value:    s,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asIntCriterionInput() (*gql.IntCriterionInput, error) {
	if c.isNullable() {
		return &gql.IntCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}

	m, ok := c.Value.(map[string]interface{})
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value, "map[string]interface{}")
	}

	v, ok := m["value"].(float64)
	if !ok {
		return nil, newUnexpectedTypeErr(m["value"], "float64")
	}

	var v2 float64
	if m["value2"] != nil {
		v2, ok = m["value2"].(float64)
		if !ok {
			return nil, newUnexpectedTypeErr(m["value2"], "float64")
		}
	}
	return &gql.IntCriterionInput{
		Value:    int(v),
		Value2:   int(v2),
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asBool() (bool, error) {
	s, ok := c.Value.(string)
	if !ok {
		return false, newUnexpectedTypeErr(c.Value, "string")
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("failed to parse bool from '%s': %w", s, err)
	}
	return b, nil
}

func (c jsonCriterion) asPHashDuplicationCriterionInput() (*gql.PHashDuplicationCriterionInput, error) {
	s, ok := c.Value.(string)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value, "string")
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bool from '%s': %w", s, err)
	}
	return &gql.PHashDuplicationCriterionInput{
		Duplicated: b,
	}, nil
}

func (c jsonCriterion) asResolutionCriterionInput() (*gql.ResolutionCriterionInput, error) {
	s, ok := c.Value.(string)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value, "string")
	}

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
	}, nil
}

func (c jsonCriterion) asString() (string, error) {
	s, ok := c.Value.(string)
	if !ok {
		return "", newUnexpectedTypeErr(c.Value, "string")
	}
	return s, nil
}

func (c jsonCriterion) asMultiCriterionInput() (*gql.MultiCriterionInput, error) {
	cs, ok := c.Value.([]interface{})
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value, "[]interface{}")
	}
	ss := make([]string, len(cs))
	for i, v := range cs {
		s, ok := v.(string)
		if !ok {
			return nil, newUnexpectedTypeErr(v, "string")
		}
		ss[i] = s
	}
	return &gql.MultiCriterionInput{
		Value:    ss,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) isNullable() bool {
	return c.Value == nil && (c.Modifier == string(gql.CriterionModifierIsNull) || c.Modifier == string(gql.CriterionModifierNotNull))
}
