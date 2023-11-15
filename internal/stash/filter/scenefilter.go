package filter

import (
	"context"
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

	sceneFilter, err := parseObjectFilter(ctx, savedFilter.Object_filter)
	if err != nil {
		return Filter{}, err
	}

	return Filter{
		FilterOpts: gql.FindFilterType{
			Direction: savedFilter.Find_filter.Direction,
			Per_page:  -1,
			Sort:      savedFilter.Find_filter.Sort,
		},
		SceneFilter: sceneFilter,
	}, nil

}

func parseObjectFilter(ctx context.Context, objects map[string]any) (gql.SceneFilterType, error) {
	var sft gql.SceneFilterType
	for k, v := range objects {
		err := setSceneFilterCriterion(ctx, k, v.(map[string]any), &sft)
		if err != nil {
			return gql.SceneFilterType{}, err
		}
	}
	return sft, nil
}

func setSceneFilterCriterion(ctx context.Context, criterionType string, criterionValue map[string]any, sceneFilter *gql.SceneFilterType) error {
	var err error
	switch criterionType {
	//HierarchicalMultiCriterionInput
	case "tags":
		sceneFilter.Tags = parseHierarchicalMultiCriterionInput(criterionValue)
	case "studios":
		sceneFilter.Studios = parseHierarchicalMultiCriterionInput(criterionValue)
	case "performer_tags":
		sceneFilter.Performer_tags = parseHierarchicalMultiCriterionInput(criterionValue)

	//StringCriterionInput
	case "title":
		sceneFilter.Title = parseStringCriterionInput(criterionValue)
	case "code":
		sceneFilter.Code = parseStringCriterionInput(criterionValue)
	case "details":
		sceneFilter.Details = parseStringCriterionInput(criterionValue)
	case "director":
		sceneFilter.Director = parseStringCriterionInput(criterionValue)
	case "oshash":
		sceneFilter.Oshash = parseStringCriterionInput(criterionValue)
	case "phash":
		sceneFilter.Phash = parseStringCriterionInput(criterionValue)
	case "path":
		sceneFilter.Path = parseStringCriterionInput(criterionValue)
	case "stash_id":
		sceneFilter.Stash_id = parseStringCriterionInput(criterionValue)
	case "url":
		sceneFilter.Url = parseStringCriterionInput(criterionValue)
	case "captions":
		sceneFilter.Captions = parseStringCriterionInput(criterionValue)
	case "audio_codec":
		sceneFilter.Audio_codec = parseStringCriterionInput(criterionValue)
	case "video_codec":
		sceneFilter.Video_codec = parseStringCriterionInput(criterionValue)
	case "checksum":
		sceneFilter.Checksum = parseStringCriterionInput(criterionValue)

	//IntCriterionInput
	case "id":
		sceneFilter.Id = parseIntCriterionInput(criterionValue)
	case "rating":
		sceneFilter.Rating = parseIntCriterionInput(criterionValue)
	case "rating100":
		sceneFilter.Rating100 = parseIntCriterionInput(criterionValue)
	case "o_counter":
		sceneFilter.O_counter = parseIntCriterionInput(criterionValue)
	case "duration":
		sceneFilter.Duration = parseIntCriterionInput(criterionValue)
	case "tag_count":
		sceneFilter.Tag_count = parseIntCriterionInput(criterionValue)
	case "performer_age":
		sceneFilter.Performer_age = parseIntCriterionInput(criterionValue)
	case "performer_count":
		sceneFilter.Performer_count = parseIntCriterionInput(criterionValue)
	case "interactive_speed":
		sceneFilter.Interactive_speed = parseIntCriterionInput(criterionValue)
	case "file_count":
		sceneFilter.File_count = parseIntCriterionInput(criterionValue)
	case "resume_time":
		sceneFilter.Resume_time = parseIntCriterionInput(criterionValue)
	case "play_count":
		sceneFilter.Play_count = parseIntCriterionInput(criterionValue)
	case "play_duration":
		sceneFilter.Play_duration = parseIntCriterionInput(criterionValue)

	//bool
	case "organized":
		err = decodeSimple(criterionValue, &sceneFilter.Organized)
	case "performer_favorite":
		err = decodeSimple(criterionValue, &sceneFilter.Performer_favorite)
	case "interactive":
		err = decodeSimple(criterionValue, &sceneFilter.Interactive)

	//PHashDuplicationCriterionInput
	case "duplicated":
		sceneFilter.Duplicated = parsePHashDuplicationCriterionInput(criterionValue)

	//PhashDistanceCriterionInput
	case "phash_distance":
		sceneFilter.Phash_distance = parsePhashDistanceCriterionInput(criterionValue)

	//ResolutionCriterionInput
	case "resolution":
		sceneFilter.Resolution = parseResolutionCriterionInput(criterionValue)

	//string
	case "has_markers":
		err = decodeSimple(criterionValue, &sceneFilter.Has_markers)
	case "is_missing":
		err = decodeSimple(criterionValue, &sceneFilter.Is_missing)

	//MultiCriterionInput
	case "movies":
		sceneFilter.Movies = parseMultiCriterionInput(criterionValue)
	case "performers":
		sceneFilter.Performers = parseMultiCriterionInput(criterionValue)

	//TimestampCriterionInput
	case "created_at":
		sceneFilter.Created_at = parseTimestampCriterionInput(criterionValue)
	case "updated_at":
		sceneFilter.Updated_at = parseTimestampCriterionInput(criterionValue)

	//DateCriterionInput
	case "date":
		sceneFilter.Date = parseDateCriterionInput(criterionValue)

	//StashIDCriterionInput
	case "stash_id_endpoint":
		sceneFilter.Stash_id_endpoint = parseStashIDCriterionInput(criterionValue)

	default:
		log.Ctx(ctx).Warn().Str("type", criterionType).Interface("value", criterionValue).Msg("Ignoring unsupported criterion")
	}
	if err != nil {
		return fmt.Errorf("failed to parse criterion (%v): %w", criterionType, err)
	}
	return nil
}
