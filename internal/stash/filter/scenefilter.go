package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
)

type Filter struct {
	FilterOpts  gql.FindFilterType
	SceneFilter gql.SceneFilterType
}

func SavedFilterToSceneFilter(ctx context.Context, savedFilter gql.SavedFilterParts) (Filter, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return Filter{}, fmt.Errorf("unsupported filter mode")
	}

	filterQuery, err := parseJsonEncodedFilter(ctx, savedFilter.Filter)
	if err != nil {
		return Filter{}, fmt.Errorf("parseJsonEncodedFilter: %w", err)
	}
	return filterQuery, nil
}

func parseJsonEncodedFilter(ctx context.Context, raw string) (Filter, error) {
	var filter jsonFilter

	err := json.Unmarshal([]byte(raw), &filter)
	if err != nil {
		return Filter{}, fmt.Errorf("unmarshal json scene filter '%s': %w", raw, err)
	}
	f, err := parseSceneFilterCriteria(ctx, filter.C)
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

func parseSceneFilterCriteria(ctx context.Context, jsonCriteria []string) (gql.SceneFilterType, error) {
	f := gql.SceneFilterType{}
	for _, jsonCriterion := range jsonCriteria {
		c, err := parseJsonCriterion(jsonCriterion)
		if err != nil {
			return gql.SceneFilterType{}, fmt.Errorf("parseJsonCriterion: %w", err)
		}
		err = setSceneFilterCriterion(ctx, c, &f)
		if err != nil {
			return gql.SceneFilterType{}, fmt.Errorf("setSceneFilterCriterion: %w", err)
		}
	}
	return f, nil
}

func setSceneFilterCriterion(ctx context.Context, criterion jsonCriterion, sceneFilter *gql.SceneFilterType) error {
	var err error
	switch criterion.Type {
	//HierarchicalMultiCriterionInput
	case "tags":
		sceneFilter.Tags, err = criterion.asHierarchicalMultiCriterionInput()
	case "studios":
		sceneFilter.Studios, err = criterion.asHierarchicalMultiCriterionInput()
	case "performerTags":
		sceneFilter.Performer_tags, err = criterion.asHierarchicalMultiCriterionInput()

	//StringCriterionInput
	case "title":
		sceneFilter.Title, err = criterion.asStringCriterionInput()
	case "code":
		sceneFilter.Code, err = criterion.asStringCriterionInput()
	case "details":
		sceneFilter.Details, err = criterion.asStringCriterionInput()
	case "director":
		sceneFilter.Director, err = criterion.asStringCriterionInput()
	case "oshash":
		sceneFilter.Oshash, err = criterion.asStringCriterionInput()
	case "sceneChecksum":
		sceneFilter.Checksum, err = criterion.asStringCriterionInput()
	case "phash":
		sceneFilter.Phash, err = criterion.asStringCriterionInput()
	case "path":
		sceneFilter.Path, err = criterion.asStringCriterionInput()
	case "stash_id":
		sceneFilter.Stash_id, err = criterion.asStringCriterionInput()
	case "url":
		sceneFilter.Url, err = criterion.asStringCriterionInput()
	case "captions":
		sceneFilter.Captions, err = criterion.asStringCriterionInput()

	//IntCriterionInput
	case "id":
		sceneFilter.Id, err = criterion.asIntCriterionInput()
	case "rating":
		sceneFilter.Rating, err = criterion.asIntCriterionInput()
	case "rating100":
		sceneFilter.Rating100, err = criterion.asIntCriterionInput()
	case "o_counter":
		sceneFilter.O_counter, err = criterion.asIntCriterionInput()
	case "duration":
		sceneFilter.Duration, err = criterion.asIntCriterionInput()
	case "tag_count":
		sceneFilter.Tag_count, err = criterion.asIntCriterionInput()
	case "performer_age":
		sceneFilter.Performer_age, err = criterion.asIntCriterionInput()
	case "performer_count":
		sceneFilter.Performer_count, err = criterion.asIntCriterionInput()
	case "interactive_speed":
		sceneFilter.Interactive_speed, err = criterion.asIntCriterionInput()
	case "file_count":
		sceneFilter.File_count, err = criterion.asIntCriterionInput()
	case "resume_time":
		sceneFilter.Resume_time, err = criterion.asIntCriterionInput()
	case "play_count":
		sceneFilter.Play_count, err = criterion.asIntCriterionInput()
	case "play_duration":
		sceneFilter.Play_duration, err = criterion.asIntCriterionInput()

	//bool
	case "organized":
		sceneFilter.Organized, err = criterion.asBool()
	case "performer_favorite":
		sceneFilter.Performer_favorite, err = criterion.asBool()
	case "interactive":
		sceneFilter.Interactive, err = criterion.asBool()

	//PHashDuplicationCriterionInput
	case "duplicated":
		sceneFilter.Duplicated, err = criterion.asPHashDuplicationCriterionInput()

	//ResolutionCriterionInput
	case "resolution":
		sceneFilter.Resolution, err = criterion.asResolutionCriterionInput()

	//string
	case "hasMarkers":
		sceneFilter.Has_markers, err = criterion.asString()
	case "sceneIsMissing":
		sceneFilter.Is_missing, err = criterion.asString()

	//MultiCriterionInput
	case "movies":
		sceneFilter.Movies, err = criterion.asMultiCriterionInput()
	case "performers":
		sceneFilter.Performers, err = criterion.asMultiCriterionInput()

	//TimestampCriterionInput
	case "created_at":
		sceneFilter.Created_at, err = criterion.asTimestampCriterionInput()
	case "updated_at":
		sceneFilter.Updated_at, err = criterion.asTimestampCriterionInput()

	//DateCriterionInput
	case "date":
		sceneFilter.Date, err = criterion.asDateCriterionInput()
	case "stash_id_endpoint":
		sceneFilter.Stash_id_endpoint, err = criterion.asStashIDCriterionInput()
		//StashIDCriterionInput

	default:
		log.Ctx(ctx).Warn().Str("type", criterion.Type).Interface("value", criterion.Value).Msg("Ignoring unsupported criterion")
	}
	if err != nil {
		return fmt.Errorf("failed to parse criterion (%v): %w", criterion, err)
	}
	return nil
}
