package filter

import (
	"errors"
	"fmt"
	"slices"
	"stash-vr/internal/stash/gql"
	"strconv"
	"strings"
)

// Saved filters created in the Stash UI store criteria in the UI's own format, e.g.
//
//	{"modifier": "INCLUDES", "value": {"items": [{"id": "1", "label": "a"}], "excluded": [], "depth": 0}}
//	{"modifier": "EQUALS", "value": {"value": 0}}
//	{"modifier": "EQUALS", "value": "4k"}
//
// Stash also accepts and stores filters saved through the saveFilter GraphQL mutation,
// which typically use the SceneFilterType input shape instead, e.g.
//
//	{"modifier": "INCLUDES", "value": ["1"], "excludes": [], "depth": 0}
//	{"modifier": "EQUALS", "value": 0}
//	{"modifier": "EQUALS", "value": "FOUR_K"}
//
// Both shapes are supported. Anything else results in an error.

// parseSimple reads a boolean or string criterion. The Stash UI stores {"modifier": "EQUALS", "value": "true"},
// the GraphQL API shape is the bare value.
func parseSimple[T string | bool](v any) (*T, error) {
	if c, ok := v.(map[string]any); ok {
		v = c["value"]
	}
	s, ok := scalarString(v)
	if !ok {
		return nil, fmt.Errorf("unexpected value %v", v)
	}
	var out T
	switch p := any(&out).(type) {
	case *string:
		*p = s
	case *bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("unexpected value %q", s)
		}
		*p = b
	}
	return &out, nil
}

func object(v any) (map[string]any, error) {
	c, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected an object, got %v", v)
	}
	return c, nil
}

// criterion returns the object and modifier of a criterion with a modifier.
func criterion(v any) (map[string]any, gql.CriterionModifier, error) {
	c, err := object(v)
	if err != nil {
		return nil, "", err
	}
	m, ok := c["modifier"].(string)
	if !ok {
		return nil, "", errors.New("missing modifier")
	}
	return c, gql.CriterionModifier(m), nil
}

// isNullModifier reports whether the modifier takes no value.
func isNullModifier(m gql.CriterionModifier) bool {
	return m == gql.CriterionModifierIsNull || m == gql.CriterionModifierNotNull
}

// valueObject returns the object holding the fields of a criterion: "value" in the UI shape,
// the criterion itself in the GraphQL API shape.
func valueObject(c map[string]any) map[string]any {
	if v, ok := c["value"].(map[string]any); ok {
		return v
	}
	return c
}

func scalarString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(x), true
	}
	return "", false
}

func stringValue(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	s, ok := scalarString(v)
	if !ok {
		return nil, fmt.Errorf("expected a string, got %v", v)
	}
	return &s, nil
}

func requiredString(v any) (string, error) {
	s, err := stringValue(v)
	if err != nil {
		return "", err
	}
	if s == nil {
		return "", errors.New("missing value")
	}
	return *s, nil
}

func intValue(v any) (*int, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case float64:
		i := int(x)
		return &i, nil
	case int:
		return &x, nil
	}
	return nil, fmt.Errorf("expected a number, got %v", v)
}

// parseIDs reads a list of ids, given either as {"id": "1", "label": "a"} objects (UI) or as bare ids (GraphQL API).
func parseIDs(v any) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	list, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected a list, got %v", v)
	}
	ids := make([]string, len(list))
	for i, item := range list {
		if o, ok := item.(map[string]any); ok {
			item = o["id"]
		}
		switch id := item.(type) {
		case string:
			ids[i] = id
		case float64:
			ids[i] = strconv.Itoa(int(id))
		default:
			return nil, fmt.Errorf("unexpected id %v", item)
		}
	}
	return ids, nil
}

// parseMultiValue returns the ids, excluded ids and depth of a (hierarchical) multi criterion.
// Some UI criteria store the list of items directly in "value", same as the GraphQL API shape.
func parseMultiValue(c map[string]any) (value []string, excludes []string, depth *int, err error) {
	var items, excluded any
	if v, ok := c["value"].(map[string]any); ok {
		items, excluded, depth = v["items"], v["excluded"], Get[int](v, "depth")
	} else {
		items, excluded, depth = c["value"], c["excludes"], Get[int](c, "depth")
	}
	if value, err = parseIDs(items); err != nil {
		return nil, nil, nil, err
	}
	if excludes, err = parseIDs(excluded); err != nil {
		return nil, nil, nil, err
	}
	return value, excludes, depth, nil
}

func parseIntCriterionInput(v any) (*gql.IntCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.IntCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	o := valueObject(c)
	value, err := intValue(o["value"])
	if err != nil {
		return nil, err
	}
	if value != nil {
		out.Value = *value
	}
	if out.Value2, err = intValue(o["value2"]); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseHierarchicalMultiCriterionInput(v any) (*gql.HierarchicalMultiCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.HierarchicalMultiCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	if out.Value, out.Excludes, out.Depth, err = parseMultiValue(c); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseMultiCriterionInput(v any) (*gql.MultiCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.MultiCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	if out.Value, out.Excludes, _, err = parseMultiValue(c); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseTimestampCriterionInput(v any) (*gql.TimestampCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.TimestampCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	o := valueObject(c)
	if out.Value, err = requiredString(o["value"]); err != nil {
		return nil, err
	}
	if out.Value2, err = stringValue(o["value2"]); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseDateCriterionInput(v any) (*gql.DateCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.DateCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	o := valueObject(c)
	if out.Value, err = requiredString(o["value"]); err != nil {
		return nil, err
	}
	if out.Value2, err = stringValue(o["value2"]); err != nil {
		return nil, err
	}
	return &out, nil
}

func parsePhashDistanceCriterionInput(v any) (*gql.PhashDistanceCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.PhashDistanceCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	o := valueObject(c)
	if out.Value, err = requiredString(o["value"]); err != nil {
		return nil, err
	}
	if out.Distance, err = intValue(o["distance"]); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseResolutionCriterionInput(v any) (*gql.ResolutionCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.ResolutionCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}

	value, err := requiredString(c["value"])
	if err != nil {
		return nil, err
	}
	switch value {
	case "144p":
		out.Value = gql.ResolutionEnumVeryLow
	case "240p":
		out.Value = gql.ResolutionEnumLow
	case "360p":
		out.Value = gql.ResolutionEnumR360p
	case "480p":
		out.Value = gql.ResolutionEnumStandard
	case "540p":
		out.Value = gql.ResolutionEnumWebHd
	case "720p":
		out.Value = gql.ResolutionEnumStandardHd
	case "1080p":
		out.Value = gql.ResolutionEnumFullHd
	case "1440p":
		out.Value = gql.ResolutionEnumQuadHd
	case "1920p":
		out.Value = gql.ResolutionEnumVrHd
	case "4k":
		out.Value = gql.ResolutionEnumFourK
	case "5k":
		out.Value = gql.ResolutionEnumFiveK
	case "6k":
		out.Value = gql.ResolutionEnumSixK
	case "7k":
		out.Value = gql.ResolutionEnumSevenK
	case "8k":
		out.Value = gql.ResolutionEnumEightK
	case "Huge":
		out.Value = gql.ResolutionEnumHuge
	default:
		// GraphQL API shape: the enum name, e.g. FOUR_K.
		if !slices.Contains(gql.AllResolutionEnum, gql.ResolutionEnum(value)) {
			return nil, fmt.Errorf("unknown resolution %q", value)
		}
		out.Value = gql.ResolutionEnum(value)
	}

	return &out, nil
}

func parseStashIDCriterionInput(v any) (*gql.StashIDCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.StashIDCriterionInput{
		Modifier: mod,
	}
	o := valueObject(c)
	if out.Endpoint, err = stringValue(o["endpoint"]); err != nil {
		return nil, err
	}
	stashID := o["stashID"]
	if stashID == nil {
		stashID = o["stash_id"]
	}
	if out.Stash_id, err = stringValue(stashID); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseDuplicationCriterionInput(v any) (*gql.DuplicationCriterionInput, error) {
	c, err := object(v)
	if err != nil {
		return nil, err
	}
	o := valueObject(c)
	out := gql.DuplicationCriterionInput{}

	out.Duplicated = Get[bool](o, "duplicated")
	out.Distance = Get[int](o, "distance")
	out.Phash = Get[bool](o, "phash")
	out.Stash_id = Get[bool](o, "stash_id")
	out.Title = Get[bool](o, "title")
	out.Url = Get[bool](o, "url")

	return &out, nil
}

func parseStringCriterionInput(v any) (*gql.StringCriterionInput, error) {
	c, mod, err := criterion(v)
	if err != nil {
		return nil, err
	}
	out := gql.StringCriterionInput{
		Modifier: mod,
	}
	if isNullModifier(out.Modifier) {
		return &out, nil
	}
	if out.Value, err = requiredString(c["value"]); err != nil {
		return nil, err
	}
	return &out, nil
}

func parseCaptionCriterionInput(v any) (*gql.StringCriterionInput, error) {
	out, err := parseStringCriterionInput(v)
	if err != nil || isNullModifier(out.Modifier) {
		return out, err
	}
	switch out.Value {
	case "Deutsche":
		out.Value = "de"
	case "English":
		out.Value = "en"
	case "Español":
		out.Value = "es"
	case "Français":
		out.Value = "fr"
	case "Italiano":
		out.Value = "it"
	case "日本":
		out.Value = "ja"
	case "한국인":
		out.Value = "ko"
	case "Holandés":
		out.Value = "nl"
	case "Português":
		out.Value = "pt"
	case "Русский":
		out.Value = "ru"
	case "Unknown":
		out.Value = "00"
	}
	// Any other value is kept as is: the GraphQL API shape stores the language code.
	return out, nil
}

func parseOrientationCriterionInput(v any) (*gql.OrientationCriterionInput, error) {
	c, err := object(v)
	if err != nil {
		return nil, err
	}
	values, ok := c["value"].([]any)
	if !ok {
		return nil, fmt.Errorf("expected a list, got %v", c["value"])
	}
	out := gql.OrientationCriterionInput{
		Value: make([]gql.OrientationEnum, len(values)),
	}
	for i, x := range values {
		s, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected orientation %v", x)
		}
		out.Value[i] = gql.OrientationEnum(strings.ToUpper(s))
	}
	return &out, nil
}
