package heresphere

import (
	"encoding/json"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/util"
)

type httpHandler struct {
	Client graphql.Client
}

func (h *httpHandler) indexHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := util.GetBaseUrl(req)

	data := buildIndex(ctx, h.Client, baseUrl)

	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) scanHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := util.GetBaseUrl(req)

	data, err := buildScan(ctx, h.Client, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("scan")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) videoDataHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("body: read")
		return
	}

	var updateVideoData updateVideoData
	err = json.Unmarshal(body, &updateVideoData)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Bytes("body", body).Msg("body: unmarshal")
	} else {
		if updateVideoData.isUpdateRequest() {
			update(ctx, h.Client, sceneId, updateVideoData)
			w.WriteHeader(http.StatusOK)
			return
		}

		if updateVideoData.isDeleteRequest() {
			destroy(ctx, h.Client, sceneId)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	data, err := buildVideoData(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}

}
