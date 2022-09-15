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
	sceneId := chi.URLParam(req, "sceneId")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("videodata: body: read")
		return
	}

	var updateVideoData UpdateVideoData
	err = json.Unmarshal(body, &updateVideoData)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("id", sceneId).Bytes("body", body).Msg("videodata: body: unmarshal")
	} else {
		if updateVideoData.IsUpdateRequest() {
			update(ctx, h.Client, sceneId, updateVideoData)
			w.WriteHeader(http.StatusOK)
			return
		}

		if updateVideoData.IsDeleteRequest() {
			if err := destroy(ctx, h.Client, sceneId); err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("sceneId", sceneId).Msg("Failed to fulfill delete request")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	videoData, err := buildVideoData(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", sceneId).Msg("buildVideoData")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.Write(ctx, w, videoData); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", sceneId).Msg("videodata: write response")
	}

}
