package library

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strconv"
	"time"
)

func (service *Service) GetScenes(ctx context.Context) (map[string]*VideoData, error) {
	res, err, _ := service.single.Do("scenes", func() (interface{}, error) {
		start := time.Now()
		service.mu.RLock()
		toFetch := make([]int, 0, len(service.vdCache))
		for k, vd := range service.vdCache {
			if vd == nil {
				id, _ := strconv.Atoi(k)
				toFetch = append(toFetch, id)
			}
		}
		service.mu.RUnlock()

		if len(toFetch) > 0 {
			resp, err := gql.FindScenes(ctx, service.stashClient, toFetch)
			if err != nil {
				return nil, fmt.Errorf("FindScenes: %w", err)
			}

			service.mu.Lock()
			for _, s := range resp.FindScenes.Scenes {
				service.vdCache[s.Id] = &VideoData{SceneParts: &s.SceneParts}
			}
			service.mu.Unlock()
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
		service.mu.RLock()
		vd := service.vdCache[id]
		service.mu.RUnlock()
		if vd != nil {
			log.Ctx(ctx).Trace().Str("id", id).Msg("Return scene from cache")
			return vd, nil
		}
	}
	vd, err := service.fetchVideoData(ctx, id)
	if err != nil {
		return nil, err
	}
	service.mu.Lock()
	service.vdCache[id] = vd
	service.mu.Unlock()
	log.Ctx(ctx).Trace().Str("id", id).Msg("Return scene from fetch")
	return vd, nil
}

func (service *Service) fetchVideoData(ctx context.Context, id string) (*VideoData, error) {
	iid, _ := strconv.Atoi(id)
	sceneIds := []int{iid}
	resp, err := gql.FindScenes(ctx, service.stashClient, sceneIds)
	if err != nil {
		return nil, fmt.Errorf("FindScenes: %w", err)
	}
	vd := VideoData{SceneParts: &resp.FindScenes.Scenes[0].SceneParts}
	return &vd, nil
}
