package deovr

import (
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"net/http"
	"stash-vr/internal/util"
)

type HttpHandler struct {
	Client graphql.Client
}

func (h HttpHandler) Index(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := fmt.Sprintf("http://%s", req.Host)

	index, err := buildIndex(ctx, h.Client, baseUrl)
	if err != nil {
		log.Error().Err(err).Msg("buildIndex")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := write(w, index); err != nil {
		log.Error().Err(err).Str("handler", "index").Msg("write")
	}
}

func (h HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	videoId := chi.URLParam(req, "videoId")

	videoData, err := buildVideoData(ctx, h.Client, videoId)
	if err != nil {
		log.Error().Str("videoId", videoId).Err(err).Msg("buildVideoData")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := write(w, videoData); err != nil {
		log.Error().Err(err).Str("handler", "videodata").Msg("write")
	}
}

func write(w http.ResponseWriter, data interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	err := util.NewJsonEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
