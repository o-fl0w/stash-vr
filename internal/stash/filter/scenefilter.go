package filter

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strings"
)

type Filter struct {
	FilterOpts  gql.FindFilterType
	SceneFilter gql.SceneFilterType
}

var perPage = -1

func SavedFilterToSceneFilter(ctx context.Context, savedFilter gql.SavedFilterParts) (Filter, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return Filter{}, fmt.Errorf("unsupported filter mode (%s)", savedFilter.Mode)
	}

	sceneFilter, err := parseObjectFilter(ctx, *savedFilter.Object_filter)
	if err != nil {
		return Filter{}, err
	}

	if savedFilter.Find_filter.Sort != nil && strings.HasPrefix(*savedFilter.Find_filter.Sort, "random_") {
		*savedFilter.Find_filter.Sort = "random"
	}

	filter := Filter{
		FilterOpts: gql.FindFilterType{
			Direction: savedFilter.Find_filter.Direction,
			Per_page:  &perPage,
			Sort:      savedFilter.Find_filter.Sort,
		},
		SceneFilter: sceneFilter,
	}

	return filter, nil

}

func parseObjectFilter(ctx context.Context, objects map[string]any) (gql.SceneFilterType, error) {
	var sft gql.SceneFilterType
	for k, v := range objects {
		setSceneFilterCriterion(ctx, k, v.(map[string]any), &sft)
	}
	return sft, nil
}

func setSceneFilterCriterion(ctx context.Context, criterionType string, criterionValue map[string]any, sceneFilter *gql.SceneFilterType) {
	switch criterionType {
	case "audio_codec":
		sceneFilter.Audio_codec = parseStringCriterionInput(criterionValue)
	case "bitrate":
		sceneFilter.Bitrate = parseIntCriterionInput(criterionValue)
	case "captions":
		sceneFilter.Captions = parseCaptionCriterionInput(criterionValue)
	case "checksum":
		sceneFilter.Checksum = parseStringCriterionInput(criterionValue)
	case "code":
		sceneFilter.Code = parseStringCriterionInput(criterionValue)
	case "created_at":
		sceneFilter.Created_at = parseTimestampCriterionInput(criterionValue)
	case "date":
		sceneFilter.Date = parseDateCriterionInput(criterionValue)
	case "details":
		sceneFilter.Details = parseStringCriterionInput(criterionValue)
	case "director":
		sceneFilter.Director = parseStringCriterionInput(criterionValue)
	case "duplicated":
		sceneFilter.Duplicated = parsePHashDuplicationCriterionInput(criterionValue)
	case "duration":
		sceneFilter.Duration = parseIntCriterionInput(criterionValue)
	case "file_count":
		sceneFilter.File_count = parseIntCriterionInput(criterionValue)
	case "framerate":
		sceneFilter.Framerate = parseIntCriterionInput(criterionValue)
	case "galleries":
		sceneFilter.Galleries = parseMultiCriterionInput(criterionValue)
	case "groups":
		sceneFilter.Groups = parseHierarchicalMultiCriterionInput(criterionValue)
	case "has_markers":
		decodeSimple(criterionValue, &sceneFilter.Has_markers)
	case "id":
		sceneFilter.Id = parseIntCriterionInput(criterionValue)
	case "interactive":
		decodeSimple(criterionValue, &sceneFilter.Interactive)
	case "interactive_speed":
		sceneFilter.Interactive_speed = parseIntCriterionInput(criterionValue)
	case "is_missing":
		decodeSimple(criterionValue, &sceneFilter.Is_missing)
	case "last_played_at":
		sceneFilter.Last_played_at = parseTimestampCriterionInput(criterionValue)
	case "movies":
		sceneFilter.Movies = parseMultiCriterionInput(criterionValue)
	case "o_counter":
		sceneFilter.O_counter = parseIntCriterionInput(criterionValue)
	case "organized":
		decodeSimple(criterionValue, &sceneFilter.Organized)
	case "orientation":
		sceneFilter.Orientation = parseOrientationCriterionInput(criterionValue)
	case "oshash":
		sceneFilter.Oshash = parseStringCriterionInput(criterionValue)
	case "path":
		sceneFilter.Path = parseStringCriterionInput(criterionValue)
	case "performer_age":
		sceneFilter.Performer_age = parseIntCriterionInput(criterionValue)
	case "performer_count":
		sceneFilter.Performer_count = parseIntCriterionInput(criterionValue)
	case "performer_favorite":
		decodeSimple(criterionValue, &sceneFilter.Performer_favorite)
	case "performer_tags":
		sceneFilter.Performer_tags = parseHierarchicalMultiCriterionInput(criterionValue)
	case "performers":
		sceneFilter.Performers = parseMultiCriterionInput(criterionValue)
	case "phash":
		sceneFilter.Phash = parseStringCriterionInput(criterionValue)
	case "phash_distance":
		sceneFilter.Phash_distance = parsePhashDistanceCriterionInput(criterionValue)
	case "play_count":
		sceneFilter.Play_count = parseIntCriterionInput(criterionValue)
	case "play_duration":
		sceneFilter.Play_duration = parseIntCriterionInput(criterionValue)
	case "rating100":
		sceneFilter.Rating100 = parseIntCriterionInput(criterionValue)
	case "resolution":
		sceneFilter.Resolution = parseResolutionCriterionInput(criterionValue)
	case "resume_time":
		sceneFilter.Resume_time = parseIntCriterionInput(criterionValue)
	case "stash_id_endpoint":
		sceneFilter.Stash_id_endpoint = parseStashIDCriterionInput(criterionValue)
	case "studios":
		sceneFilter.Studios = parseHierarchicalMultiCriterionInput(criterionValue)
	case "tag_count":
		sceneFilter.Tag_count = parseIntCriterionInput(criterionValue)
	case "tags":
		sceneFilter.Tags = parseHierarchicalMultiCriterionInput(criterionValue)
	case "title":
		sceneFilter.Title = parseStringCriterionInput(criterionValue)
	case "updated_at":
		sceneFilter.Updated_at = parseTimestampCriterionInput(criterionValue)
	case "url":
		sceneFilter.Url = parseStringCriterionInput(criterionValue)
	case "video_codec":
		sceneFilter.Video_codec = parseStringCriterionInput(criterionValue)
	default:
		log.Ctx(ctx).Debug().Str("type", criterionType).Interface("value", criterionValue).Msg("Ignoring unsupported criterion")
	}
}
