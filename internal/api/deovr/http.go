package deovr

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/util"
)

type httpHandler struct {
	Client graphql.Client
}

func (h httpHandler) indexHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := util.GetBaseUrl(req)

	data := buildIndex(ctx, h.Client, baseUrl)

	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h httpHandler) videoDataHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")

	data, err := buildVideoData(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
