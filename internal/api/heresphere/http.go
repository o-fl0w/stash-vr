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

	index := buildIndex(ctx, h.Client, baseUrl)

	if err := common.WriteJson(ctx, w, index); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("body: read")
		return
	}

	var updateVideoData UpdateVideoData
	err = json.Unmarshal(body, &updateVideoData)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Bytes("body", body).Msg("body: unmarshal")
	} else {
		if updateVideoData.IsUpdateRequest() {
			update(ctx, h.Client, sceneId, updateVideoData)
			w.WriteHeader(http.StatusOK)
			return
		}

		if updateVideoData.IsDeleteRequest() {
			if err := destroy(ctx, h.Client, sceneId); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("Delete failed")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	videoData, err := buildVideoData(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := common.WriteJson(ctx, w, videoData); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}

}
