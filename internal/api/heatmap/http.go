package heatmap

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"image/jpeg"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/library"
	"stash-vr/internal/stash"
)

func CoverHandler(libraryService *library.Service) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sceneId := chi.URLParam(r, "videoId")

		vd, err := libraryService.GetScene(ctx, sceneId, false)
		if err != nil {
			log.Ctx(ctx).Debug().Msg("Scene not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		p := vd.SceneParts.Paths
		cover, err := buildHeatmapCover(ctx, stash.ApiKeyed(*p.Screenshot), stash.ApiKeyed(*p.Interactive_heatmap))
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("buildHeatmapCover")
			if errors.Is(err, errImageNotFound) {
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
	return baseUrl + "/cover/" + sceneId
}
