package heresphere

import (
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
	"stash-vr/internal/stimhub"
)

type httpHandler struct {
	StashClient   graphql.Client
	StimhubClient *stimhub.Client
}

func (h *httpHandler) indexHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)

	data := buildIndex(ctx, h.StashClient, h.StimhubClient, baseUrl)

	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) scanHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := internal.GetBaseUrl(req)

	data, err := buildScan(ctx, h.StashClient, h.StimhubClient, baseUrl)
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
	baseUrl := internal.GetBaseUrl(req)
	videoId, err := url.QueryUnescape(chi.URLParam(req, "videoId"))
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("malformed videoId")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sceneId, _, _ := stimhub.SplitStimSceneId(videoId)

	vdReq, err := internal.UnmarshalBody[videoDataRequest](req)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to parse request body")
		//w.WriteHeader(http.StatusBadRequest)
		//return
	}

	if vdReq.isUpdateRequest() {
		update(ctx, h.StashClient, sceneId, vdReq)
		w.WriteHeader(http.StatusOK)
		return
	}

	if vdReq.isDeleteRequest() {
		destroy(ctx, h.StashClient, sceneId)
		w.WriteHeader(http.StatusOK)
		return
	}

	if vdReq.isPlayRequest() && !config.Get().IsPlayCountDisabled {
		incrementPlayCount(ctx, h.StashClient, sceneId)
	}

	var includeMediaSource = vdReq.NeedsMediaSource == nil || *vdReq.NeedsMediaSource

	data, err := buildVideoData(ctx, h.StashClient, h.StimhubClient, baseUrl, videoId, includeMediaSource)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *httpHandler) videoHspHandler(hspDir string) http.HandlerFunc {
	filesDir := http.Dir(hspDir)
	return func(w http.ResponseWriter, r *http.Request) {
		sceneId := chi.URLParam(r, "videoId")
		r.URL.Path = sceneId + ".hsp"
		fs := http.FileServer(filesDir)
		fs.ServeHTTP(w, r)
	}
}

func (h *httpHandler) eventsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	event, err := internal.UnmarshalBody[playbackEvent](req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to parse request body")
		w.WriteHeader(http.StatusBadRequest)
	}

	log.Ctx(ctx).Trace().Str("event", fmt.Sprintf("%v", event)).Send()
	return
}
