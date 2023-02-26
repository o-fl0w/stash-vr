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
	Modifier string `json:"modifier"`
	Type     string `json:"type"`
	Value    any    `json:"value"`
}

type errUnexpectedType struct {
	source any
}

type stringAnyMap = map[string]any

func (e errUnexpectedType) Error() string {
	return fmt.Sprintf("could not assert '%T' with value='%v'", e.source, e.source)
}

func newUnexpectedTypeErr(source any) *errUnexpectedType {
	return &errUnexpectedType{source}
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
	if c.Value == nil {
		return &gql.HierarchicalMultiCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}
	m, ok := c.Value.(stringAnyMap)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}

	items, err := getValue[[]any](m, "items")
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(items))
	for i, item := range items {
		mid, ok := item.(stringAnyMap)
		if !ok {
			return nil, newUnexpectedTypeErr(item)
		}
		id, ok := mid["id"].(string)
		if !ok {
			return nil, newUnexpectedTypeErr(mid["id"])
		}
		ids[i] = id
	}

	return &gql.HierarchicalMultiCriterionInput{
		Value:    ids,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asStringCriterionInput() (*gql.StringCriterionInput, error) {
	if c.Value == nil {
		return &gql.StringCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}
	s, ok := c.Value.(string)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}
	return &gql.StringCriterionInput{
		Value:    s,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asIntCriterionInput() (*gql.IntCriterionInput, error) {
	if c.Value == nil {
		return &gql.IntCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}
	m, ok := c.Value.(stringAnyMap)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}

	value, err := getValue[float64](m, "value")
	if err != nil {
		return nil, newUnexpectedTypeErr(m["value"])
	}

	value2, err := getValue[float64](m, "value2")
	if err != nil {
		return nil, newUnexpectedTypeErr(m["value2"])
	}

	return &gql.IntCriterionInput{
		Value:    int(value),
		Value2:   int(value2),
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asBool() (bool, error) {
	value, ok := c.Value.(string)
	if !ok {
		return false, newUnexpectedTypeErr(c.Value)
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("failed to parse bool from '%s': %w", value, err)
	}
	return b, nil
}

func (c jsonCriterion) asPHashDuplicationCriterionInput() (*gql.PHashDuplicationCriterionInput, error) {
	b, err := c.asBool()
	if err != nil {
		return nil, err
	}
	return &gql.PHashDuplicationCriterionInput{
		Duplicated: b,
	}, nil
}

func (c jsonCriterion) asResolutionCriterionInput() (*gql.ResolutionCriterionInput, error) {
	if c.Value == nil {
		return &gql.ResolutionCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}
	value, ok := c.Value.(string)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}

	var res gql.ResolutionEnum

	switch value {
	case "144p":
		res = gql.ResolutionEnumVeryLow
	case "240p":
		res = gql.ResolutionEnumLow
	case "360p":
		res = gql.ResolutionEnumR360p
	case "480p":
		res = gql.ResolutionEnumStandard
	case "540p":
		res = gql.ResolutionEnumWebHd
	case "720p":
		res = gql.ResolutionEnumStandardHd
	case "1080p":
		res = gql.ResolutionEnumFullHd
	case "1440p":
		res = gql.ResolutionEnumQuadHd
	case "1920p":
		res = gql.ResolutionEnumVrHd
	case "4k":
		res = gql.ResolutionEnumFourK
	case "5k":
		res = gql.ResolutionEnumFiveK
	case "6k":
		res = gql.ResolutionEnumSixK
	case "8k":
		res = gql.ResolutionEnumEightK
	}

	return &gql.ResolutionCriterionInput{
		Value:    res,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asString() (string, error) {
	s, ok := c.Value.(string)
	if !ok {
		return "", newUnexpectedTypeErr(c.Value)
	}
	return s, nil
}

func (c jsonCriterion) asMultiCriterionInput() (*gql.MultiCriterionInput, error) {
	if c.Value == nil {
		return &gql.MultiCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}

	value, ok := c.Value.([]any)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}
	ss := make([]string, len(value))
	for i, v := range value {
		s, ok := v.(string)
		if !ok {
			return nil, newUnexpectedTypeErr(v)
		}
		ss[i] = s
	}
	return &gql.MultiCriterionInput{
		Value:    ss,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asTimestampCriterionInput() (*gql.TimestampCriterionInput, error) {
	if c.Value == nil {
		return &gql.TimestampCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}

	m, ok := c.Value.(stringAnyMap)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}

	value, err := getValue[string](m, "value")
	if err != nil {
		return nil, err
	}

	value2, err := getValue[string](m, "value2")
	if err != nil {
		return nil, err
	}

	return &gql.TimestampCriterionInput{
		Value:    value,
		Value2:   value2,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asDateCriterionInput() (*gql.DateCriterionInput, error) {
	if c.Value == nil {
		return &gql.DateCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}

	m, ok := c.Value.(stringAnyMap)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}

	value, err := getValue[string](m, "value")
	if err != nil {
		return nil, err
	}

	value2, err := getValue[string](m, "value2")
	if err != nil {
		return nil, err
	}

	return &gql.DateCriterionInput{
		Value:    value,
		Value2:   value2,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func (c jsonCriterion) asStashIDCriterionInput() (*gql.StashIDCriterionInput, error) {
	if c.Value == nil {
		return &gql.StashIDCriterionInput{
			Modifier: gql.CriterionModifier(c.Modifier),
		}, nil
	}

	m, ok := c.Value.(stringAnyMap)
	if !ok {
		return nil, newUnexpectedTypeErr(c.Value)
	}

	endpoint, err := getValue[string](m, "endpoint")
	if err != nil {
		return nil, err
	}

	stashId, err := getValue[string](m, "stash_id")
	if err != nil {
		return nil, err
	}

	return &gql.StashIDCriterionInput{
		Endpoint: endpoint,
		Stash_id: stashId,
		Modifier: gql.CriterionModifier(c.Modifier),
	}, nil
}

func getValue[T any](m stringAnyMap, key string) (T, error) {
	_, ok := m[key]
	if !ok {
		return *new(T), nil
	}
	value, ok := m[key].(T)
	if !ok {
		return *new(T), newUnexpectedTypeErr(m[key])
	}
	return value, nil
}
