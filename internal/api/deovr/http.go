package deovr

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/library"
)

type httpHandler struct {
	LibraryService *library.Service
}

func (h httpHandler) indexHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)

	sections, err := h.LibraryService.GetSections(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get sections")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vds, err := h.LibraryService.GetScenes(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get scenes")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dto, err := buildIndex(sections, vds, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to build index")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := internal.WriteJson(ctx, w, dto); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error writing response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h httpHandler) videoDataHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")
	baseUrl := internal.GetBaseUrl(req)

	vd, err := h.LibraryService.GetScene(ctx, sceneId, false)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get scene data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dto, err := buildVideoData(vd, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to build video data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := internal.WriteJson(ctx, w, dto); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error writing response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
