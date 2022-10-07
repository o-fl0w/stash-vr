package funscript

import (
	"errors"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

func CoverHandler(client graphql.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sceneId := chi.URLParam(r, "videoId")

		response, err := gql.FindHeatmapCoverBySceneId(ctx, client, sceneId)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("FindHeatmapCoverBySceneId")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		p := response.FindScene.Paths
		cover, err := GetHeatmapCover(ctx, stash.ApiKeyed(p.Screenshot), stash.ApiKeyed(p.Interactive_heatmap))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("GetHeatmapCover")
			if errors.Is(err, NotFoundErr) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		_, err = w.Write(cover)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("cover: write")
			return
		}
	}
}
