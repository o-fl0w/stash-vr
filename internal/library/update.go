package library

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"slices"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
	"time"
)

func (libraryService *Service) UpdateRating(ctx context.Context, id string, rating5 float32) error {
	newRating100 := int(rating5 * 20)

	_, err := gql.SceneUpdateRating100(ctx, libraryService.StashClient, id, &newRating100)
	if err != nil {
		return fmt.Errorf("SceneUpdateRating100: %w", err)
	}
	return nil
}

func (libraryService *Service) UpdateFavorite(ctx context.Context, id string, isFavoriteRequested bool) error {
	favoriteTagName := config.Application().FavoriteTag

	if favoriteTagName == "" {
		log.Ctx(ctx).Info().Msg("Sync favorite requested but FAVORITE_TAG is empty, ignoring request")
		return nil
	}

	favoriteTagId, err := stash.FindOrCreateTag(ctx, libraryService.StashClient, favoriteTagName)
	if err != nil {
		return err
	}

	response, err := gql.FindSceneTags(ctx, libraryService.StashClient, id)
	if err != nil {
		return fmt.Errorf("FindSceneTags: %w", err)
	}

	newTagIds := make([]string, 0, len(response.FindScene.Tags)+1)

	var hasFavoriteTag bool
	for _, t := range response.FindScene.Tags {
		if t.Id == favoriteTagId {
			hasFavoriteTag = true
			if !isFavoriteRequested {
				continue
			}
		}
		newTagIds = append(newTagIds, t.Id)
	}
	if !hasFavoriteTag && isFavoriteRequested {
		newTagIds = append(newTagIds, favoriteTagId)
	}

	if _, err := gql.SceneUpdateTags(ctx, libraryService.StashClient, id, newTagIds); err != nil {
		return fmt.Errorf("SceneUpdateTags: %w", err)
	}

	return nil
}

func (libraryService *Service) UpdateTags(ctx context.Context, id string, tags []string) error {
	tagIds := make([]string, len(tags))
	for i, tag := range tags {
		tagId, err := stash.FindOrCreateTag(ctx, libraryService.StashClient, tag)
		if err != nil {
			return err
		}
		tagIds[i] = tagId
	}
	if _, err := gql.SceneUpdateTags(ctx, libraryService.StashClient, id, tagIds); err != nil {
		return fmt.Errorf("SceneUpdateTags: %w", err)
	}
	return nil
}

type MarkerDto struct {
	PrimaryTagName string
	StartSecond    float64
	EndSecond      *float64
	Title          string
	MarkerId       string //hack: use the rating field for transport of marker id
}

func (libraryService *Service) UpdateMarkers(ctx context.Context, id string, incomingMarkers []MarkerDto) error {
	vd, err := libraryService.GetScene(ctx, id, false)
	if err != nil {
		return err
	}

	markersToDestroy := make([]string, 0)
	for _, existingMarker := range vd.SceneParts.Scene_markers {
		if !slices.ContainsFunc(incomingMarkers, func(m MarkerDto) bool {
			return m.MarkerId == existingMarker.Id
		}) {
			markersToDestroy = append(markersToDestroy, existingMarker.Id)
		}
	}

	markersToUpdate := make([]MarkerDto, 0)
	markersToCreate := make([]MarkerDto, 0)

	for _, incoming := range incomingMarkers {
		if incoming.MarkerId != "" && incoming.MarkerId != "0" && slices.ContainsFunc(vd.SceneParts.Scene_markers, func(existingMarker *gql.ScenePartsScene_markersSceneMarker) bool {
			return incoming.MarkerId == existingMarker.Id
		}) {
			markersToUpdate = append(markersToUpdate, incoming)
		} else {
			markersToCreate = append(markersToCreate, incoming)
		}
	}

	for _, m := range markersToUpdate {
		tagId, err := stash.FindOrCreateTag(ctx, libraryService.StashClient, m.PrimaryTagName)
		if err != nil {
			return fmt.Errorf("failed to find or create primary tag for marker: %w", err)
		}
		_, err = gql.SceneMarkerUpdate(ctx, libraryService.StashClient, m.MarkerId, tagId, m.StartSecond, m.EndSecond, m.Title)
		if err != nil {
			return fmt.Errorf("SceneMarkerCreate: %w", err)
		}
	}
	for _, m := range markersToCreate {
		tagId, err := stash.FindOrCreateTag(ctx, libraryService.StashClient, m.PrimaryTagName)
		if err != nil {
			return fmt.Errorf("failed to find or create primary tag for marker: %w", err)
		}
		_, err = gql.SceneMarkerCreate(ctx, libraryService.StashClient, id, tagId, m.StartSecond, m.EndSecond, m.Title)
		if err != nil {
			return fmt.Errorf("SceneMarkerCreate: %w", err)
		}
	}

	_, err = gql.SceneMarkersDestroy(ctx, libraryService.StashClient, markersToDestroy)
	if err != nil {
		return fmt.Errorf("SceneMarkersDestroy: %w", err)
	}

	return nil
}

func (libraryService *Service) ClearAndCreateMarkers(ctx context.Context, id string, markers []MarkerDto) error {
	resp, err := gql.FindSceneMarkers(ctx, libraryService.StashClient, id)
	if err != nil {
		return fmt.Errorf("FindSceneMarkers: %w", err)
	}
	currentMarkers := make([]MarkerDto, len(resp.FindSceneMarkers.Scene_markers))
	for i, m := range resp.FindSceneMarkers.Scene_markers {
		currentMarkers[i] = MarkerDto{
			PrimaryTagName: m.Primary_tag.Name,
			StartSecond:    m.Seconds * 1000,
			Title:          m.Title,
		}
		if m.End_seconds != nil {
			currentMarkers[i].EndSecond = util.Ptr(*m.End_seconds * 1000)
		}
	}
	if util.UnorderedEqual(currentMarkers, markers) {
		return nil
	}
	markersToDestroy := make([]string, len(resp.FindSceneMarkers.Scene_markers))
	for i, sm := range resp.FindSceneMarkers.Scene_markers {
		markersToDestroy[i] = sm.Id
	}
	_, err = gql.SceneMarkersDestroy(ctx, libraryService.StashClient, markersToDestroy)
	if err != nil {
		return fmt.Errorf("SceneMarkersDestroy: %w", err)
	}

	for _, m := range markers {
		tagId, err := stash.FindOrCreateTag(ctx, libraryService.StashClient, m.PrimaryTagName)
		if err != nil {
			return fmt.Errorf("failed to find or create primary tag for marker: %w", err)
		}
		_, err = gql.SceneMarkerCreate(ctx, libraryService.StashClient, id, tagId, m.StartSecond, m.EndSecond, m.Title)
		if err != nil {
			return fmt.Errorf("SceneMarkerCreate: %w", err)
		}
	}
	return nil
}

func (libraryService *Service) Delete(ctx context.Context, id string) error {
	if _, err := gql.SceneDestroy(ctx, libraryService.StashClient, id); err != nil {
		return fmt.Errorf("SceneDestroy: %w", err)
	}
	return nil
}

func (libraryService *Service) IncrementO(ctx context.Context, id string) error {
	_, err := gql.SceneIncrementO(ctx, libraryService.StashClient, id)
	if err != nil {
		return fmt.Errorf("SceneIncrementO: %w", err)
	}
	return nil
}

func (libraryService *Service) DecrementO(ctx context.Context, id string) error {
	_, err := gql.SceneDecrementO(ctx, libraryService.StashClient, id)
	if err != nil {
		return fmt.Errorf("SceneDecrementO: %w", err)
	}
	return nil
}

func (libraryService *Service) IncrementPlayCount(ctx context.Context, id string) error {
	_, err := gql.SceneIncrementPlayCount(ctx, libraryService.StashClient, id)
	if err != nil {
		return fmt.Errorf("SceneIncrementPlayCount: %w", err)
	}
	return nil
}

func (libraryService *Service) DecrementPlayCount(ctx context.Context, id string) error {
	_, err := gql.SceneDecrementPlayCount(ctx, libraryService.StashClient, id)
	if err != nil {
		return fmt.Errorf("SceneDecrementPlayCount: %w", err)
	}
	return nil
}

func (libraryService *Service) SetOrganized(ctx context.Context, id string, newState bool) error {
	_, err := gql.SceneUpdateOrganized(ctx, libraryService.StashClient, id, &newState)
	if err != nil {
		return fmt.Errorf("SceneUpdateOrganized: %w", err)
	}
	return nil
}

func (libraryService *Service) AddPlayDuration(ctx context.Context, id string, duration time.Duration) error {
	seconds := duration.Seconds()
	_, err := gql.SceneAddPlayDurationSeconds(ctx, libraryService.StashClient, id, &seconds)
	if err != nil {
		return fmt.Errorf("SceneAddPlayDurationSeconds: %w", err)
	}
	return nil
}
