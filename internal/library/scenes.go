package library

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"maps"
	"stash-vr/internal/stash/gql"
	"strconv"
	"time"
)

func (service *Service) GetScenes(ctx context.Context) (map[string]*VideoData, error) {
	res, err, _ := service.single.Do("scenes", func() (interface{}, error) {
		start := time.Now()
		service.muVdCache.RLock()
		toFetch := make([]int, 0, len(service.vdCache))
		for k, vd := range service.vdCache {
			if vd == nil {
				id, _ := strconv.Atoi(k)
				toFetch = append(toFetch, id)
			}
		}
		service.muVdCache.RUnlock()

		if len(toFetch) > 0 {
			resp, err := gql.FindScenes(ctx, service.StashClient, toFetch)
			if err != nil {
				return nil, fmt.Errorf("FindScenes: %w", err)
			}

			service.muVdCache.Lock()
			for _, s := range resp.FindScenes.Scenes {
				vd := VideoData{SceneParts: &s.SceneParts}
				service.decorateTags(ctx, &vd)
				service.vdCache[s.Id] = &vd
			}
			service.muVdCache.Unlock()
			elapsed := time.Since(start)
			log.Ctx(ctx).Trace().Int("fetched", len(toFetch)).Dur("ms", elapsed).Msg("Updated cache")
		} else {
			log.Ctx(ctx).Trace().Msg("Cache hit, no scenes to fetch")
		}
		return service.snapshot(), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(map[string]*VideoData), nil
}

func (service *Service) GetScene(ctx context.Context, id string, forceFetch bool) (*VideoData, error) {
	if !forceFetch {
		service.muVdCache.RLock()
		vd := service.vdCache[id]
		service.muVdCache.RUnlock()
		if vd != nil {
			log.Ctx(ctx).Trace().Str("id", id).Msg("Return scene from cache")
			return vd, nil
		}
	}
	vd, err := service.fetchVideoData(ctx, id)
	if err != nil {
		return nil, err
	}
	service.decorateTags(ctx, vd)
	service.muVdCache.Lock()
	service.vdCache[id] = vd
	service.muVdCache.Unlock()
	log.Ctx(ctx).Trace().Str("id", id).Msg("Return scene from fetch")
	return vd, nil
}

func (service *Service) fetchVideoData(ctx context.Context, id string) (*VideoData, error) {
	iid, _ := strconv.Atoi(id)
	sceneIds := []int{iid}
	resp, err := gql.FindScenes(ctx, service.StashClient, sceneIds)
	if err != nil {
		return nil, fmt.Errorf("FindScenes: %w", err)
	}
	vd := VideoData{SceneParts: &resp.FindScenes.Scenes[0].SceneParts}
	return &vd, nil
}

func (service *Service) decorateTags(ctx context.Context, vd *VideoData) {
	allAncestors := map[string]struct{}{}
	for _, t := range vd.SceneParts.Tags {
		ancestors := service.ancestorTagNames(ctx, t.Id)
		for _, a := range ancestors {
			allAncestors[a] = struct{}{}
		}
	}
	for a := range maps.Keys(allAncestors) {
		vd.SceneParts.Tags = append(vd.SceneParts.Tags, &gql.TagPartsArrayTagsTag{TagParts: gql.TagParts{
			Id:   "",
			Name: "#" + a},
		})
	}
}
