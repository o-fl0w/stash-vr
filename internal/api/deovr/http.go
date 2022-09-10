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

	index, err := buildIndex(ctx, h.Client, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("buildIndex")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := common.Write(ctx, w, index); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("handler", "index").Msg("write")
	}
}

func (h HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	videoId := chi.URLParam(req, "videoId")

	videoData, err := buildVideoData(ctx, h.Client, videoId)
	if err != nil {
		log.Ctx(ctx).Error().Str("videoId", videoId).Err(err).Msg("buildVideoData")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.Write(ctx, w, videoData); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("handler", "videodata").Msg("write")
	}
}
