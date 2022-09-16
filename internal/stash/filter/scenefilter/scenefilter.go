package scenefilter

import (
	"encoding/json"
	"fmt"
	"stash-vr/internal/stash/filter/internal"
	"stash-vr/internal/stash/gql"
)

type Filter struct {
	FilterOpts  gql.FindFilterType
	SceneFilter gql.SceneFilterType
}

func ParseJsonEncodedFilter(raw string) (Filter, error) {
	var filter internal.JsonFilter

	err := json.Unmarshal([]byte(raw), &filter)
	if err != nil {
		return Filter{}, fmt.Errorf("unmarshal json scene filter '%s': %w", raw, err)
	}
	f, err := parseSceneFilterCriteria(filter.C)
	if err != nil {
		return Filter{}, fmt.Errorf("parseSceneFilterCriteria: %w", err)
	}

	sortDir := gql.SortDirectionEnumAsc
	if filter.SortDir == "desc" {
		sortDir = gql.SortDirectionEnumDesc
	}
	return Filter{FilterOpts: gql.FindFilterType{
		Per_page:  -1,
		Sort:      filter.SortBy,
		Direction: sortDir,
	}, SceneFilter: f}, nil
}

func parseSceneFilterCriteria(jsonCriteria []string) (gql.SceneFilterType, error) {
	f := gql.SceneFilterType{}
	for _, jsonCriterion := range jsonCriteria {
		c, err := internal.ParseJsonCriterion(jsonCriterion)
		if err != nil {
			return gql.SceneFilterType{}, fmt.Errorf("parseJsonCriterion: %w", err)
		}
		setSceneFilterCriterion(c, &f)
	}
	return f, nil
}

func setSceneFilterCriterion(criterion internal.JsonCriterion, sceneFilter *gql.SceneFilterType) {
	switch criterion.Type {
	//HierarchicalMultiCriterionInput
	case "tags":
		sceneFilter.Tags = criterion.AsHierarchicalMultiCriterionInput()
	case "studios":
		sceneFilter.Studios = criterion.AsHierarchicalMultiCriterionInput()
	case "performerTags":
		sceneFilter.Performer_tags = criterion.AsHierarchicalMultiCriterionInput()

	//StringCriterionInput
	case "title":
		sceneFilter.Title = criterion.AsStringCriterionInput()
	case "details":
		sceneFilter.Details = criterion.AsStringCriterionInput()
	case "oshash":
		sceneFilter.Oshash = criterion.AsStringCriterionInput()
	case "sceneChecksum":
		sceneFilter.Checksum = criterion.AsStringCriterionInput()
	case "phash":
		sceneFilter.Phash = criterion.AsStringCriterionInput()
	case "path":
		sceneFilter.Path = criterion.AsStringCriterionInput()
	case "stash_id":
		sceneFilter.Stash_id = criterion.AsStringCriterionInput()
	case "url":
		sceneFilter.Url = criterion.AsStringCriterionInput()
	case "captions":
		sceneFilter.Captions = criterion.AsStringCriterionInput()

	//IntCriterionInput
	case "rating":
		sceneFilter.Rating = criterion.AsIntCriterionInput()
	case "o_counter":
		sceneFilter.O_counter = criterion.AsIntCriterionInput()
	case "duration":
		sceneFilter.Duration = criterion.AsIntCriterionInput()
	case "tag_count":
		sceneFilter.Tag_count = criterion.AsIntCriterionInput()
	case "performer_age":
		sceneFilter.Performer_age = criterion.AsIntCriterionInput()
	case "performer_count":
		sceneFilter.Performer_count = criterion.AsIntCriterionInput()
	case "interactive_speed":
		sceneFilter.Interactive_speed = criterion.AsIntCriterionInput()

	//bool
	case "organized":
		sceneFilter.Organized = criterion.AsBool()
	case "performer_favorite":
		sceneFilter.Performer_favorite = criterion.AsBool()
	case "interactive":
		sceneFilter.Interactive = criterion.AsBool()

	//PHashDuplicationCriterionInput
	case "duplicated":
		sceneFilter.Duplicated = criterion.AsPHashDuplicationCriterionInput()

	//ResolutionCriterionInput
	case "resolution":
		sceneFilter.Resolution = criterion.AsResolutionCriterionInput()

	//string
	case "hasMarkers":
		sceneFilter.Has_markers = criterion.AsString()
	case "sceneIsMissing":
		sceneFilter.Is_missing = criterion.AsString()

	//MultiCriterionInput
	case "movies":
		sceneFilter.Movies = criterion.AsMultiCriterionInput()
	case "performers":
		sceneFilter.Performers = criterion.AsMultiCriterionInput()
	}
}
