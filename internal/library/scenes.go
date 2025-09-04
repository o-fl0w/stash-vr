package library

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/stash/gql"
	"strconv"
	"time"
)

func (libraryService *Service) GetScenes(ctx context.Context) (map[string]*VideoData, error) {
	res, err, _ := libraryService.single.Do("scenes", func() (interface{}, error) {
		start := time.Now()
		libraryService.muVdCache.RLock()
		toFetch := make([]int, 0, len(libraryService.vdCache))
		for k, vd := range libraryService.vdCache {
			if vd == nil {
				id, _ := strconv.Atoi(k)
				toFetch = append(toFetch, id)
			}
		}
		libraryService.muVdCache.RUnlock()

		if len(toFetch) > 0 {
			vds, err := libraryService.fetchVideoData(ctx, toFetch)
			if err != nil {
				return nil, err
			}

			libraryService.muVdCache.Lock()
			for _, vd := range vds {
				libraryService.vdCache[vd.Id()] = vd
			}
			libraryService.muVdCache.Unlock()
			elapsed := time.Since(start)
			log.Ctx(ctx).Trace().Int("fetched", len(toFetch)).Dur("ms", elapsed).Msg("Updated cache")
		} else {
			log.Ctx(ctx).Trace().Msg("Cache hit, no scenes to fetch")
		}
		return libraryService.snapshot(), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(map[string]*VideoData), nil
}

func (libraryService *Service) GetScene(ctx context.Context, id string, forceFetch bool) (*VideoData, error) {
	if !forceFetch {
		libraryService.muVdCache.RLock()
		vd := libraryService.vdCache[id]
		libraryService.muVdCache.RUnlock()
		if vd != nil {
			log.Ctx(ctx).Trace().Str("id", id).Msg("Return scene from cache")
			return vd, nil
		}
	}
	iid, _ := strconv.Atoi(id)
	vds, err := libraryService.fetchVideoData(ctx, []int{iid})
	if err != nil {
		return nil, err
	}

	libraryService.muVdCache.Lock()
	libraryService.vdCache[id] = vds[0]
	libraryService.muVdCache.Unlock()
	log.Ctx(ctx).Trace().Str("id", id).Msg("Return scene from fetch")
	return vds[0], nil
}

func (libraryService *Service) fetchVideoData(ctx context.Context, sceneIds []int) ([]*VideoData, error) {
	resp, err := gql.FindScenes(ctx, libraryService.StashClient, sceneIds)
	if err != nil {
		return nil, fmt.Errorf("FindScenes: %w", err)
	}
	vds := make([]*VideoData, len(resp.FindScenes.Scenes))
	for i, s := range resp.FindScenes.Scenes {
		vd := VideoData{SceneParts: &s.SceneParts}
		libraryService.decorateTags(&vd)
		vds[i] = &vd
	}
	return vds, nil
}
