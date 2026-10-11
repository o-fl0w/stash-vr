package filter

import (
	"context"
	"fmt"
	"stash-vr/internal/stash/gql"
	"strings"

	"github.com/rs/zerolog/log"
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

	var objectFilter map[string]any
	if savedFilter.Object_filter != nil {
		objectFilter = *savedFilter.Object_filter
	}
	sceneFilter, err := parseObjectFilter(ctx, objectFilter)
	if err != nil {
		return Filter{}, err
	}

	var sort *string
	var direction *gql.SortDirectionEnum
	if savedFilter.Find_filter != nil {
		sort, direction = savedFilter.Find_filter.Sort, savedFilter.Find_filter.Direction
	}

	if sort != nil && strings.HasPrefix(*sort, "random_") {
		*sort = "random"
	}

	filter := Filter{
		FilterOpts: gql.FindFilterType{
			Direction: direction,
			Per_page:  &perPage,
			Sort:      sort,
		},
		SceneFilter: sceneFilter,
	}

	return filter, nil

}

func parseObjectFilter(ctx context.Context, objects map[string]any) (gql.SceneFilterType, error) {
	var sft gql.SceneFilterType
	for k, v := range objects {
		if v == nil {
			// null means the criterion is not set
			continue
		}
		if err := setSceneFilterCriterion(ctx, k, v, &sft); err != nil {
			return gql.SceneFilterType{}, fmt.Errorf("criterion %s: %w", k, err)
		}
	}
	return sft, nil
}

func setSceneFilterCriterion(ctx context.Context, criterionType string, criterionValue any, sceneFilter *gql.SceneFilterType) error {
	var err error
	switch criterionType {
	case "audio_codec":
		sceneFilter.Audio_codec, err = parseStringCriterionInput(criterionValue)
	case "bitrate":
		sceneFilter.Bitrate, err = parseIntCriterionInput(criterionValue)
	case "captions":
		sceneFilter.Captions, err = parseCaptionCriterionInput(criterionValue)
	case "checksum":
		sceneFilter.Checksum, err = parseStringCriterionInput(criterionValue)
	case "code":
		sceneFilter.Code, err = parseStringCriterionInput(criterionValue)
	case "created_at":
		sceneFilter.Created_at, err = parseTimestampCriterionInput(criterionValue)
	case "date":
		sceneFilter.Date, err = parseDateCriterionInput(criterionValue)
	case "details":
		sceneFilter.Details, err = parseStringCriterionInput(criterionValue)
	case "director":
		sceneFilter.Director, err = parseStringCriterionInput(criterionValue)
	case "duplicated":
		sceneFilter.Duplicated, err = parseDuplicationCriterionInput(criterionValue)
	case "duration":
		sceneFilter.Duration, err = parseIntCriterionInput(criterionValue)
	case "file_count":
		sceneFilter.File_count, err = parseIntCriterionInput(criterionValue)
	case "framerate":
		sceneFilter.Framerate, err = parseIntCriterionInput(criterionValue)
	case "galleries":
		sceneFilter.Galleries, err = parseMultiCriterionInput(criterionValue)
	case "groups":
		sceneFilter.Groups, err = parseHierarchicalMultiCriterionInput(criterionValue)
	case "has_markers":
		sceneFilter.Has_markers, err = parseSimple[string](criterionValue)
	case "id":
		sceneFilter.Id, err = parseIntCriterionInput(criterionValue)
	case "interactive":
		sceneFilter.Interactive, err = parseSimple[bool](criterionValue)
	case "interactive_speed":
		sceneFilter.Interactive_speed, err = parseIntCriterionInput(criterionValue)
	case "is_missing":
		sceneFilter.Is_missing, err = parseSimple[string](criterionValue)
	case "last_played_at":
		sceneFilter.Last_played_at, err = parseTimestampCriterionInput(criterionValue)
	case "movies":
		sceneFilter.Movies, err = parseMultiCriterionInput(criterionValue)
	case "o_counter":
		sceneFilter.O_counter, err = parseIntCriterionInput(criterionValue)
	case "organized":
		sceneFilter.Organized, err = parseSimple[bool](criterionValue)
	case "orientation":
		sceneFilter.Orientation, err = parseOrientationCriterionInput(criterionValue)
	case "oshash":
		sceneFilter.Oshash, err = parseStringCriterionInput(criterionValue)
	case "path":
		sceneFilter.Path, err = parseStringCriterionInput(criterionValue)
	case "performer_age":
		sceneFilter.Performer_age, err = parseIntCriterionInput(criterionValue)
	case "performer_count":
		sceneFilter.Performer_count, err = parseIntCriterionInput(criterionValue)
	case "performer_favorite":
		sceneFilter.Performer_favorite, err = parseSimple[bool](criterionValue)
	case "performer_tags":
		sceneFilter.Performer_tags, err = parseHierarchicalMultiCriterionInput(criterionValue)
	case "performers":
		sceneFilter.Performers, err = parseMultiCriterionInput(criterionValue)
	case "phash":
		sceneFilter.Phash, err = parseStringCriterionInput(criterionValue)
	case "phash_distance":
		sceneFilter.Phash_distance, err = parsePhashDistanceCriterionInput(criterionValue)
	case "play_count":
		sceneFilter.Play_count, err = parseIntCriterionInput(criterionValue)
	case "play_duration":
		sceneFilter.Play_duration, err = parseIntCriterionInput(criterionValue)
	case "rating100":
		sceneFilter.Rating100, err = parseIntCriterionInput(criterionValue)
	case "resolution":
		sceneFilter.Resolution, err = parseResolutionCriterionInput(criterionValue)
	case "resume_time":
		sceneFilter.Resume_time, err = parseIntCriterionInput(criterionValue)
	case "stash_id_endpoint":
		sceneFilter.Stash_id_endpoint, err = parseStashIDCriterionInput(criterionValue)
	case "studios":
		sceneFilter.Studios, err = parseHierarchicalMultiCriterionInput(criterionValue)
	case "tag_count":
		sceneFilter.Tag_count, err = parseIntCriterionInput(criterionValue)
	case "tags":
		sceneFilter.Tags, err = parseHierarchicalMultiCriterionInput(criterionValue)
	case "title":
		sceneFilter.Title, err = parseStringCriterionInput(criterionValue)
	case "updated_at":
		sceneFilter.Updated_at, err = parseTimestampCriterionInput(criterionValue)
	case "url":
		sceneFilter.Url, err = parseStringCriterionInput(criterionValue)
	case "video_codec":
		sceneFilter.Video_codec, err = parseStringCriterionInput(criterionValue)
	default:
		log.Ctx(ctx).Debug().Str("type", criterionType).Interface("value", criterionValue).Msg("Ignoring unsupported criterion")
	}
	return err
}
