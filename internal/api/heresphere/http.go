package heresphere

import (
	"context"
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

func (h *HttpHandler) Index(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := fmt.Sprintf("http://%s", req.Host)

	index, err := buildIndex(ctx, h.Client, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("buildIndex")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := common.Write(ctx, w, index); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("index: write")
	}
}

func (h *HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	videoId := chi.URLParam(req, "videoId")

	readVideoData(ctx, w, h.Client, videoId)
}

func readVideoData(ctx context.Context, w http.ResponseWriter, client graphql.Client, videoId string) {
	videoData, err := buildVideoData(ctx, client, videoId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", videoId).Msg("buildVideoData")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.Write(ctx, w, videoData); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", videoId).Msg("videodata: read: write")
	}
}
