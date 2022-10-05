package internal

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/common"
	"stash-vr/internal/api/deovr/internal/index"
	"stash-vr/internal/api/deovr/internal/videodata"
	"stash-vr/internal/util"
)

type HttpHandler struct {
	Client graphql.Client
}

func (h HttpHandler) Index(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := util.GetBaseUrl(req)

	data := index.Build(ctx, h.Client, baseUrl)

	if err := common.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")

	data, err := videodata.Build(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
