package filter

import (
	"context"
	"encoding/json"
	"reflect"
	"stash-vr/internal/stash/gql"
	"testing"
)

func ptr[T any](v T) *T {
	return &v
}

func objectFilter(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("invalid test json: %v", err)
	}
	return m
}

func nilIfEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}

// normalize treats empty and missing id lists the same, as Stash does.
func normalize(f *gql.SceneFilterType) {
	for _, c := range []*gql.MultiCriterionInput{f.Galleries, f.Movies, f.Performers} {
		if c != nil {
			c.Value, c.Excludes = nilIfEmpty(c.Value), nilIfEmpty(c.Excludes)
		}
	}
	for _, c := range []*gql.HierarchicalMultiCriterionInput{f.Groups, f.Performer_tags, f.Studios, f.Tags} {
		if c != nil {
			c.Value, c.Excludes = nilIfEmpty(c.Value), nilIfEmpty(c.Excludes)
		}
	}
}

// Saved filter as created through the saveFilter GraphQL mutation, using the SceneFilterType shape.
const filter9API = `{
	"performers": {"modifier": "INCLUDES", "value": ["1169", "107", "2792", "1735", "2111", "2836", "108", "3128", "1573", "1307", "1851", "325", "2453", "126", "2621", "2799", "2540", "3062"]},
	"play_count": {"modifier": "EQUALS", "value": 0},
	"resolution": {"modifier": "EQUALS", "value": "FOUR_K"}
}`

// The same filter as stored by the Stash UI.
const filter9UI = `{
	"performers": {"modifier": "INCLUDES", "value": {"excluded": [], "items": [
		{"id": "1169", "label": "a"}, {"id": "107", "label": "b"}, {"id": "2792", "label": "c"},
		{"id": "1735", "label": "d"}, {"id": "2111", "label": "e"}, {"id": "2836", "label": "f"},
		{"id": "108", "label": "g"}, {"id": "3128", "label": "h"}, {"id": "1573", "label": "i"},
		{"id": "1307", "label": "j"}, {"id": "1851", "label": "k"}, {"id": "325", "label": "l"},
		{"id": "2453", "label": "m"}, {"id": "126", "label": "n"}, {"id": "2621", "label": "o"},
		{"id": "2799", "label": "p"}, {"id": "2540", "label": "q"}, {"id": "3062", "label": "r"}
	]}},
	"play_count": {"modifier": "EQUALS", "value": {"value": 0}},
	"resolution": {"modifier": "EQUALS", "value": "4k"}
}`

func TestSavedFilterToSceneFilter(t *testing.T) {
	want := Filter{
		FilterOpts: gql.FindFilterType{
			Direction: ptr(gql.SortDirectionEnumDesc),
			Per_page:  &perPage,
			Sort:      ptr("date"),
		},
		SceneFilter: gql.SceneFilterType{
			Performers: &gql.MultiCriterionInput{
				Modifier: gql.CriterionModifierIncludes,
				Value:    []string{"1169", "107", "2792", "1735", "2111", "2836", "108", "3128", "1573", "1307", "1851", "325", "2453", "126", "2621", "2799", "2540", "3062"},
			},
			Play_count: &gql.IntCriterionInput{Modifier: gql.CriterionModifierEquals, Value: 0},
			Resolution: &gql.ResolutionCriterionInput{Modifier: gql.CriterionModifierEquals, Value: gql.ResolutionEnumFourK},
		},
	}

	for name, s := range map[string]string{"api": filter9API, "ui": filter9UI} {
		t.Run(name, func(t *testing.T) {
			of := objectFilter(t, s)
			got, err := SavedFilterToSceneFilter(context.Background(), gql.SavedFilterParts{
				Id:   "9",
				Name: "2160p Unwatched",
				Mode: gql.FilterModeScenes,
				Find_filter: &gql.SavedFilterPartsFind_filterSavedFindFilterType{
					Sort:      ptr("date"),
					Direction: ptr(gql.SortDirectionEnumDesc),
				},
				Object_filter: &of,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			normalize(&got.SceneFilter)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func TestSavedFilterToSceneFilterWithoutFilters(t *testing.T) {
	got, err := SavedFilterToSceneFilter(context.Background(), gql.SavedFilterParts{Id: "1", Mode: gql.FilterModeScenes})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Filter{FilterOpts: gql.FindFilterType{Per_page: &perPage}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseObjectFilter(t *testing.T) {
	tests := []struct {
		name string
		ui   string
		api  string
		want gql.SceneFilterType
	}{
		{
			name: "hierarchical multi",
			ui:   `{"tags": {"modifier": "INCLUDES_ALL", "value": {"depth": -1, "excluded": [{"id": "3", "label": "c"}], "items": [{"id": "1", "label": "a"}, {"id": "2", "label": "b"}]}}}`,
			api:  `{"tags": {"modifier": "INCLUDES_ALL", "value": ["1", "2"], "excludes": ["3"], "depth": -1}}`,
			want: gql.SceneFilterType{Tags: &gql.HierarchicalMultiCriterionInput{
				Modifier: gql.CriterionModifierIncludesAll, Value: []string{"1", "2"}, Excludes: []string{"3"}, Depth: ptr(-1),
			}},
		},
		{
			name: "multi with item list",
			ui:   `{"galleries": {"modifier": "INCLUDES", "value": [{"id": "5", "label": "e"}]}}`,
			api:  `{"galleries": {"modifier": "INCLUDES", "value": ["5"]}}`,
			want: gql.SceneFilterType{Galleries: &gql.MultiCriterionInput{
				Modifier: gql.CriterionModifierIncludes, Value: []string{"5"},
			}},
		},
		{
			name: "multi excludes only",
			ui:   `{"performers": {"modifier": "INCLUDES", "value": {"excluded": [{"id": "7", "label": "g"}], "items": []}}}`,
			api:  `{"performers": {"modifier": "INCLUDES", "excludes": ["7"]}}`,
			want: gql.SceneFilterType{Performers: &gql.MultiCriterionInput{
				Modifier: gql.CriterionModifierIncludes, Excludes: []string{"7"},
			}},
		},
		{
			name: "int range",
			ui:   `{"duration": {"modifier": "BETWEEN", "value": {"value": 180, "value2": 600}}}`,
			api:  `{"duration": {"modifier": "BETWEEN", "value": 180, "value2": 600}}`,
			want: gql.SceneFilterType{Duration: &gql.IntCriterionInput{
				Modifier: gql.CriterionModifierBetween, Value: 180, Value2: ptr(600),
			}},
		},
		{
			name: "date and timestamp",
			ui:   `{"date": {"modifier": "BETWEEN", "value": {"value": "2020-01-01", "value2": "2021-01-01"}}, "created_at": {"modifier": "GREATER_THAN", "value": {"value": "2024-01-01 00:00"}}}`,
			api:  `{"date": {"modifier": "BETWEEN", "value": "2020-01-01", "value2": "2021-01-01"}, "created_at": {"modifier": "GREATER_THAN", "value": "2024-01-01 00:00"}}`,
			want: gql.SceneFilterType{
				Date:       &gql.DateCriterionInput{Modifier: gql.CriterionModifierBetween, Value: "2020-01-01", Value2: ptr("2021-01-01")},
				Created_at: &gql.TimestampCriterionInput{Modifier: gql.CriterionModifierGreaterThan, Value: "2024-01-01 00:00"},
			},
		},
		{
			name: "phash distance",
			ui:   `{"phash_distance": {"modifier": "EQUALS", "value": {"value": "abc", "distance": 4}}}`,
			api:  `{"phash_distance": {"modifier": "EQUALS", "value": "abc", "distance": 4}}`,
			want: gql.SceneFilterType{Phash_distance: &gql.PhashDistanceCriterionInput{
				Modifier: gql.CriterionModifierEquals, Value: "abc", Distance: ptr(4),
			}},
		},
		{
			name: "resolution",
			ui:   `{"resolution": {"modifier": "GREATER_THAN", "value": "7k"}}`,
			api:  `{"resolution": {"modifier": "GREATER_THAN", "value": "SEVEN_K"}}`,
			want: gql.SceneFilterType{Resolution: &gql.ResolutionCriterionInput{
				Modifier: gql.CriterionModifierGreaterThan, Value: gql.ResolutionEnumSevenK,
			}},
		},
		{
			name: "booleans and strings",
			ui:   `{"organized": {"modifier": "EQUALS", "value": "false"}, "interactive": {"modifier": "EQUALS", "value": "true"}, "is_missing": {"modifier": "EQUALS", "value": "title"}, "has_markers": {"modifier": "EQUALS", "value": "true"}}`,
			api:  `{"organized": false, "interactive": true, "is_missing": "title", "has_markers": "true"}`,
			want: gql.SceneFilterType{
				Organized:   ptr(false),
				Interactive: ptr(true),
				Is_missing:  ptr("title"),
				Has_markers: ptr("true"),
			},
		},
		{
			name: "string and captions",
			ui:   `{"title": {"modifier": "INCLUDES", "value": "foo"}, "captions": {"modifier": "INCLUDES", "value": "English"}}`,
			api:  `{"title": {"modifier": "INCLUDES", "value": "foo"}, "captions": {"modifier": "INCLUDES", "value": "en"}}`,
			want: gql.SceneFilterType{
				Title:    &gql.StringCriterionInput{Modifier: gql.CriterionModifierIncludes, Value: "foo"},
				Captions: &gql.StringCriterionInput{Modifier: gql.CriterionModifierIncludes, Value: "en"},
			},
		},
		{
			name: "orientation",
			ui:   `{"orientation": {"value": ["landscape"]}}`,
			api:  `{"orientation": {"value": ["LANDSCAPE"]}}`,
			want: gql.SceneFilterType{Orientation: &gql.OrientationCriterionInput{
				Value: []gql.OrientationEnum{gql.OrientationEnumLandscape},
			}},
		},
		{
			name: "stash id",
			ui:   `{"stash_id_endpoint": {"modifier": "EQUALS", "value": {"endpoint": "https://stashdb.org/graphql", "stashID": "x"}}}`,
			api:  `{"stash_id_endpoint": {"modifier": "EQUALS", "endpoint": "https://stashdb.org/graphql", "stash_id": "x"}}`,
			want: gql.SceneFilterType{Stash_id_endpoint: &gql.StashIDCriterionInput{
				Modifier: gql.CriterionModifierEquals, Endpoint: ptr("https://stashdb.org/graphql"), Stash_id: ptr("x"),
			}},
		},
		{
			name: "duplicated",
			ui:   `{"duplicated": {"value": {"phash": true}}}`,
			api:  `{"duplicated": {"phash": true}}`,
			want: gql.SceneFilterType{Duplicated: &gql.DuplicationCriterionInput{Phash: ptr(true)}},
		},
		{
			name: "null modifiers without value",
			ui:   `{"date": {"modifier": "NOT_NULL"}, "performers": {"modifier": "IS_NULL"}, "title": {"modifier": "NOT_NULL"}}`,
			api:  `{"date": {"modifier": "NOT_NULL"}, "performers": {"modifier": "IS_NULL"}, "title": {"modifier": "NOT_NULL"}, "tags": null}`,
			want: gql.SceneFilterType{
				Date:       &gql.DateCriterionInput{Modifier: gql.CriterionModifierNotNull},
				Performers: &gql.MultiCriterionInput{Modifier: gql.CriterionModifierIsNull},
				Title:      &gql.StringCriterionInput{Modifier: gql.CriterionModifierNotNull},
			},
		},
	}
	for _, tt := range tests {
		for shape, s := range map[string]string{"ui": tt.ui, "api": tt.api} {
			t.Run(tt.name+"/"+shape, func(t *testing.T) {
				got, err := parseObjectFilter(context.Background(), objectFilter(t, s))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				normalize(&got)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("got %+v, want %+v", got, tt.want)
				}
			})
		}
	}
}

func TestParseObjectFilterInvalid(t *testing.T) {
	tests := map[string]string{
		"criterion not an object":  `{"play_count": 0}`,
		"missing modifier":         `{"play_count": {"value": 0}}`,
		"int not a number":         `{"play_count": {"modifier": "EQUALS", "value": "zero"}}`,
		"id object without id":     `{"performers": {"modifier": "INCLUDES", "value": [{"label": "a"}]}}`,
		"id not a string":          `{"performers": {"modifier": "INCLUDES", "value": [true]}}`,
		"value not a list":         `{"performers": {"modifier": "INCLUDES", "value": "1"}}`,
		"excludes not a list":      `{"tags": {"modifier": "INCLUDES", "value": ["1"], "excludes": "2"}}`,
		"items not a list":         `{"tags": {"modifier": "INCLUDES", "value": {"items": "1"}}}`,
		"date without value":       `{"date": {"modifier": "EQUALS"}}`,
		"timestamp empty value":    `{"created_at": {"modifier": "GREATER_THAN", "value": {}}}`,
		"phash without value":      `{"phash_distance": {"modifier": "EQUALS", "value": {"distance": 4}}}`,
		"string without value":     `{"title": {"modifier": "INCLUDES"}}`,
		"unknown resolution":       `{"resolution": {"modifier": "EQUALS", "value": "4K"}}`,
		"resolution without value": `{"resolution": {"modifier": "EQUALS"}}`,
		"orientation not a list":   `{"orientation": {"value": "landscape"}}`,
		"orientation not strings":  `{"orientation": {"value": [1]}}`,
		"boolean not a boolean":    `{"organized": "maybe"}`,
		"boolean without value":    `{"organized": {"modifier": "EQUALS"}}`,
		"boolean list":             `{"interactive": [true]}`,
	}
	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parseObjectFilter(context.Background(), objectFilter(t, s))
			if err == nil {
				t.Errorf("expected error for %s", s)
			}
		})
	}
}
