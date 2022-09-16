package deovr

import (
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/common"
)

type HttpHandler struct {
	Client graphql.Client
}

func (h HttpHandler) Index(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := fmt.Sprintf("http://%s", req.Host)

	index := buildIndex(ctx, h.Client, baseUrl)

	if err := common.WriteJson(ctx, w, index); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")

	videoData, err := buildVideoData(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.WriteJson(ctx, w, videoData); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
