package heatmap

import (
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"image/jpeg"
	"net/http"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

func CoverHandler(client graphql.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sceneId := chi.URLParam(r, "videoId")

		response, err := gql.FindScriptDataBySceneId(ctx, client, sceneId)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("FindHeatmapCoverBySceneId")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if response.FindScene == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		p := response.FindScene.Paths
		cover, err := getHeatmapCover(ctx, stash.ApiKeyed(p.Screenshot), stash.ApiKeyed(p.Interactive_heatmap))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("getHeatmapCover")
			if errors.Is(err, NotFoundErr) {
				w.WriteHeader(http.StatusBadGateway)
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
}

func GetCoverUrl(baseUrl string, sceneId string) string {
	return fmt.Sprintf("%s/cover/%s", baseUrl, sceneId)
}
