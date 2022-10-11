package heatmap

import (
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"image/jpeg"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

func CoverHandler(client graphql.Client) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sceneId := chi.URLParam(r, "videoId")

		response, err := gql.FindScriptDataBySceneId(ctx, client, sceneId)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("FindScriptDataBySceneId")
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if response.FindScene == nil {
			log.Ctx(ctx).Debug().Msg("Scene not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		p := response.FindScene.Paths
		cover, err := buildHeatmapCover(ctx, stash.ApiKeyed(p.Screenshot), stash.ApiKeyed(p.Interactive_heatmap))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("buildHeatmapCover")
			if errors.Is(err, imageNotFoundErr) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		err = jpeg.Encode(w, cover, nil)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("cover: write")
			return
		}
	}
	return internal.LogRoute("cover", internal.LogVideoId(f))
}

func GetCoverUrl(baseUrl string, sceneId string) string {
	return fmt.Sprintf("%s/cover/%s", baseUrl, sceneId)
}
