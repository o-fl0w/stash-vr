package filter

import (
	"stash-vr/internal/stash/gql"
	"strconv"
	"strings"
)

func decodeSimple[T string | bool](c interface{}, dst **T) {
	m := c.(map[string]interface{})
	x := m["value"].(string)
	switch any(*dst).(type) {
	case *string:
		*dst = any(&x).(*T)
	case *bool:
		b, _ := strconv.ParseBool(x)
		*dst = any(&b).(*T)
	}
}

func modifier(c map[string]any) gql.CriterionModifier {
	return gql.CriterionModifier(c["modifier"].(string))
}

func parseIntCriterionInput(c map[string]any) *gql.IntCriterionInput {
	out := gql.IntCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	out.Value = GetOr[int](c, "value.value", 0)
	out.Value2 = Get[int](c, "value.value2")
	return &out
}

func parseHierarchicalMultiCriterionInput(c map[string]any) *gql.HierarchicalMultiCriterionInput {
	out := gql.HierarchicalMultiCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	out.Depth = Get[int](c, "value.depth")

	items := Get[[]any](c, "value.items")
	if items != nil {
		out.Value = make([]string, len(*items))
		for i, o := range *items {
			out.Value[i] = *Get[string](o, "id")
		}
	}

	excluded := Get[[]any](c, "value.excluded")
	if excluded != nil {
		out.Excludes = make([]string, len(*excluded))
		for i, o := range *excluded {
			out.Excludes[i] = *Get[string](o, "id")
		}
	}
	return &out
}

func parseMultiCriterionInput(c map[string]any) *gql.MultiCriterionInput {
	out := gql.MultiCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}

	excluded := Get[[]any](c, "value.excluded")
	if excluded != nil {
		out.Excludes = make([]string, len(*excluded))
		for i, o := range *excluded {
			out.Excludes[i] = *Get[string](o, "id")
		}
	}

	items := Get[[]any](c, "value.items")
	if items != nil {
		out.Value = make([]string, len(*items))
		for i, o := range *items {
			out.Value[i] = *Get[string](o, "id")
		}
	} else {
		values := Get[[]any](c, "value")
		if values != nil {
			out.Value = make([]string, len(*values))
			for i, o := range *values {
				out.Value[i] = *Get[string](o, "id")
			}
		}
	}

	return &out
}

func parseTimestampCriterionInput(c map[string]any) *gql.TimestampCriterionInput {
	out := gql.TimestampCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	out.Value = *Get[string](c, "value.value")
	out.Value2 = Get[string](c, "value.value2")
	return &out
}

func parseDateCriterionInput(c map[string]any) *gql.DateCriterionInput {
	out := gql.DateCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	out.Value = *Get[string](c, "value.value")
	out.Value2 = Get[string](c, "value.value2")
	return &out
}

func parsePhashDistanceCriterionInput(c map[string]any) *gql.PhashDistanceCriterionInput {
	out := gql.PhashDistanceCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	out.Value = *Get[string](c, "value.value")
	out.Distance = Get[int](c, "value.distance")
	return &out
}

func parseResolutionCriterionInput(c map[string]any) *gql.ResolutionCriterionInput {
	out := gql.ResolutionCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}

	switch *Get[string](c, "value") {
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
	case "8k":
		out.Value = gql.ResolutionEnumEightK
	case "Huge":
		out.Value = gql.ResolutionEnumHuge
	}

	return &out
}

func parseStashIDCriterionInput(c map[string]any) *gql.StashIDCriterionInput {
	out := gql.StashIDCriterionInput{
		Modifier: modifier(c),
	}
	out.Endpoint = Get[string](c, "value.endpoint")
	out.Stash_id = Get[string](c, "value.stashID")
	return &out
}

func parsePHashDuplicationCriterionInput(c map[string]any) *gql.PHashDuplicationCriterionInput {
	out := gql.PHashDuplicationCriterionInput{}

	duplicated := Get[string](c, "value")
	if duplicated != nil {
		d, _ := strconv.ParseBool(*duplicated)
		out.Duplicated = &d
	}

	return &out
}

func parseStringCriterionInput(c map[string]any) *gql.StringCriterionInput {
	out := gql.StringCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	out.Value = *Get[string](c, "value")
	return &out
}

func parseCaptionCriterionInput(c map[string]any) *gql.StringCriterionInput {
	out := gql.StringCriterionInput{
		Modifier: modifier(c),
	}
	if out.Modifier == gql.CriterionModifierIsNull {
		return &out
	}
	switch *Get[string](c, "value") {
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
	return &out
}

func parseOrientationCriterionInput(c map[string]any) *gql.OrientationCriterionInput {
	out := gql.OrientationCriterionInput{}

	values := Get[[]any](c, "value")
	out.Value = make([]gql.OrientationEnum, len(*values))
	for i, v := range *values {
		out.Value[i] = gql.OrientationEnum(strings.ToUpper(v.(string)))
	}
	return &out
}
