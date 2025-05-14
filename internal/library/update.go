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

func (service *Service) UpdateRating(ctx context.Context, id string, rating float32) error {
	newRating := int(rating * 20)

	_, err := gql.SceneUpdateRating100(ctx, service.stashClient, id, &newRating)
	if err != nil {
		return fmt.Errorf("SceneUpdateRating100: %w", err)
	}
	return nil
}

func (service *Service) UpdateFavorite(ctx context.Context, id string, isFavoriteRequested bool) error {
	favoriteTagName := config.Get().FavoriteTag

	if favoriteTagName == "" {
		log.Ctx(ctx).Info().Msg("Sync favorite requested but FAVORITE_TAG is empty, ignoring request")
		return nil
	}

	favoriteTagId, err := stash.FindOrCreateTag(ctx, service.stashClient, favoriteTagName)
	if err != nil {
		return err
	}

	response, err := gql.FindSceneTags(ctx, service.stashClient, id)
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

	if _, err := gql.SceneUpdateTags(ctx, service.stashClient, id, newTagIds); err != nil {
		return fmt.Errorf("SceneUpdateTags: %w", err)
	}

	return nil
}

func (service *Service) UpdateTags(ctx context.Context, id string, tags []string) error {
	tagIds := make([]string, len(tags))
	for i, tag := range tags {
		tagId, err := stash.FindOrCreateTag(ctx, service.stashClient, tag)
		if err != nil {
			return err
		}
		tagIds[i] = tagId
	}
	if _, err := gql.SceneUpdateTags(ctx, service.stashClient, id, tagIds); err != nil {
		return fmt.Errorf("SceneUpdateTags: %w", err)
	}
	log.Ctx(ctx).Debug().Interface("tagIds", tagIds).Msg("Updated tags")
	return nil
}

type MarkerDto struct {
	PrimaryTagName string
	StartSecond    float64
	EndSecond      *float64
	Title          string
	MarkerId       string //hack: use the rating field for transport of marker id
}

func (service *Service) UpdateMarkers(ctx context.Context, id string, incomingMarkers []MarkerDto) error {
	vd, err := service.GetScene(ctx, id, false)
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
		tagId, err := stash.FindOrCreateTag(ctx, service.stashClient, m.PrimaryTagName)
		if err != nil {
			return fmt.Errorf("failed to find or create primary tag for marker: %w", err)
		}
		_, err = gql.SceneMarkerUpdate(ctx, service.stashClient, m.MarkerId, tagId, m.StartSecond, m.EndSecond, m.Title)
		if err != nil {
			return fmt.Errorf("SceneMarkerCreate: %w", err)
		}
	}
	for _, m := range markersToCreate {
		tagId, err := stash.FindOrCreateTag(ctx, service.stashClient, m.PrimaryTagName)
		if err != nil {
			return fmt.Errorf("failed to find or create primary tag for marker: %w", err)
		}
		_, err = gql.SceneMarkerCreate(ctx, service.stashClient, id, tagId, m.StartSecond, m.EndSecond, m.Title)
		if err != nil {
			return fmt.Errorf("SceneMarkerCreate: %w", err)
		}
	}

	_, err = gql.SceneMarkersDestroy(ctx, service.stashClient, markersToDestroy)
	if err != nil {
		return fmt.Errorf("SceneMarkersDestroy: %w", err)
	}

	return nil
}

func (service *Service) ClearAndCreateMarkers(ctx context.Context, id string, markers []MarkerDto) error {
	resp, err := gql.FindSceneMarkers(ctx, service.stashClient, id)
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
	_, err = gql.SceneMarkersDestroy(ctx, service.stashClient, markersToDestroy)
	if err != nil {
		return fmt.Errorf("SceneMarkersDestroy: %w", err)
	}

	for _, m := range markers {
		tagId, err := stash.FindOrCreateTag(ctx, service.stashClient, m.PrimaryTagName)
		if err != nil {
			return fmt.Errorf("failed to find or create primary tag for marker: %w", err)
		}
		_, err = gql.SceneMarkerCreate(ctx, service.stashClient, id, tagId, m.StartSecond, m.EndSecond, m.Title)
		if err != nil {
			return fmt.Errorf("SceneMarkerCreate: %w", err)
		}
	}
	return nil
}

func (service *Service) Delete(ctx context.Context, id string) error {
	if _, err := gql.SceneDestroy(ctx, service.stashClient, id); err != nil {
		return fmt.Errorf("SceneDestroy: %w", err)
	}
	log.Ctx(ctx).Debug().Str("id", id).Msg("Destroy scene request sent to Stash")
	return nil
}

func (service *Service) IncrementO(ctx context.Context, id string) error {
	_, err := gql.SceneIncrementO(ctx, service.stashClient, id, time.Now())
	if err != nil {
		return fmt.Errorf("SceneIncrementO: %w", err)
	}
	return nil
}

func (service *Service) IncrementPlayCount(ctx context.Context, id string) error {
	_, err := gql.SceneIncrementPlayCount(ctx, service.stashClient, id, time.Now())
	if err != nil {
		return fmt.Errorf("SceneIncrementPlayCount: %w", err)
	}
	return nil
}
