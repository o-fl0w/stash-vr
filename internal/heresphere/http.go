package heresphere

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"stash-vr/internal/config"
)

type HttpHandler struct {
	Config config.Application
}

func (h HttpHandler) Index(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	ctx := context.Background()
	baseUrl := fmt.Sprintf("http://%s", req.Host)

	index, err := buildIndex(ctx, h.Config.StashGraphQLUrl, baseUrl)
	if err != nil {
		log.Error().Err(err).Msg("heresphere: buildIndex")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := write(w, index); err != nil {
		log.Error().Err(err).Msg("heresphere: http: index")
	}
}

func (h HttpHandler) VideoData(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := context.Background()
	videoId := params.ByName("videoId")

	videoData, err := buildVideoData(ctx, h.Config.StashGraphQLUrl, videoId)
	if err != nil {
		log.Error().Str("videoId", videoId).Err(err).Msg("heresphere: buildVideoData")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := write(w, videoData); err != nil {
		log.Error().Err(err).Msg("heresphere: http: videodata")
	}
}

func write(w http.ResponseWriter, data interface{}) error {
	w.Header().Add("HereSphere-JSON-Version", "1")
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
