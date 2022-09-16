package deovr

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := HttpHandler{Client: client}
	r := chi.NewRouter()

	r.Get("/", indexHandler(httpHandler.Index))
	r.Get("/{videoId}", videoDataHandler(httpHandler.VideoData))
	return r
}

func indexHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := log.With().Str("route", "index").Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func videoDataHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoId := chi.URLParam(r, "videoId")
		ctx := log.With().Str("route", "videoData").Str("videoId", videoId).Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
