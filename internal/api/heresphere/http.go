package heresphere

import (
	"encoding/json"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"io"
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
	defer req.Body.Close()

	ctx := req.Context()
	videoId := chi.URLParam(req, "videoId")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("videodata: body: read")
		return
	}

	var updateVideoData UpdateVideoData
	err = json.Unmarshal(body, &updateVideoData)
	if err != nil {
		log.Debug().Err(err).Str("id", videoId).Bytes("body", body).Msg("videodata: body: unmarshal")
	} else {
		if updateVideoData.IsUpdateRequest() {
			update(ctx, h.Client, videoId, updateVideoData)
			w.WriteHeader(http.StatusOK)
			return
		}

		if updateVideoData.IsDeleteRequest() {
			if err := destroy(ctx, h.Client, videoId); err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("videoId", videoId).Msg("Failed to fulfill delete request")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	videoData, err := buildVideoData(ctx, h.Client, videoId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", videoId).Msg("buildVideoData")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.Write(ctx, w, videoData); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", videoId).Msg("videodata: write response")
	}

}
