package deovr

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/library"
	"stash-vr/internal/multipart"
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
	videoId := chi.URLParam(req, "videoId")
	sceneId, fileId := multipart.ParseVideoId(videoId)
	baseUrl := internal.GetBaseUrl(req)

	vd, err := h.LibraryService.GetScene(ctx, sceneId, false)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get scene data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	activeItem := multipart.TargetItem(sceneId, vd.SceneParts.Files, fileId)

	dto, err := buildVideoData(vd, baseUrl, activeItem)
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

func (h httpHandler) playHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	videoId := chi.URLParam(req, "videoId")
	targetUrl := req.URL.Query().Get("url")

	if targetUrl == "" {
		http.Error(w, "missing stream url", http.StatusBadRequest)
		return
	}

	sceneId, fileId := multipart.ParseVideoId(videoId)

	if vd, err := h.LibraryService.GetScene(ctx, sceneId, false); err == nil {
		// If the target file isn't the primary one, update it in Stash
		if fileId != "" && len(vd.SceneParts.Files) > 0 && vd.SceneParts.Files[0].Id != fileId {
			log.Ctx(ctx).Info().Str("scene", sceneId).Str("file", fileId).Msg("Switching Primary File for Multi-part Scene")
			_ = h.LibraryService.SetPrimaryFile(ctx, sceneId, fileId)
		}
	}

	http.Redirect(w, req, targetUrl, http.StatusFound)
}
