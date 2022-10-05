package sync

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

func setMarkers(ctx context.Context, client graphql.Client, sceneId string, markers []marker) {
	if !config.Get().IsSyncMarkersAllowed {
		log.Ctx(ctx).Info().Msg("Sync markers requested but is disabled in config, ignoring request")
		return
	}
	response, err := gql.FindSceneMarkers(ctx, client, sceneId)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("setMarkers: FindSceneMarkers")
		return
	}
	for _, smt := range response.SceneMarkerTags {
		for _, sm := range smt.Scene_markers {
			if _, err := gql.SceneMarkerDestroy(ctx, client, sm.Id); err != nil {
				log.Ctx(ctx).Warn().Err(err).
					Str("id", sm.Id).Str("title", sm.Title).Str("tag", sm.Primary_tag.Name).Msg("setMarkers: SceneMarkerDestroy")
				continue
			}
			log.Ctx(ctx).Debug().Str("id", sm.Id).Str("title", sm.Title).Str("tag", sm.Primary_tag.Name).Msg("Marker deleted")
		}
	}
	for _, m := range markers {
		tagId, err := stash.FindOrCreateTag(ctx, client, m.tag)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("title", m.title).Str("tag", m.tag).Msg("setMarkers: FindOrCreateTag")
			continue
		}
		createResponse, err := gql.SceneMarkerCreate(ctx, client, sceneId, tagId, m.start, m.title)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("tagId", tagId).Interface("marker", m).Msg("setMarkers: SceneMarkerCreate")
			continue
		}
		log.Ctx(ctx).Debug().Str("id", createResponse.SceneMarkerCreate.Id).Str("title", m.title).Str("tag", m.tag).Msg("Marker created")
	}
}
