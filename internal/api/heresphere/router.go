package heresphere

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/heresphere/internal"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := internal.HttpHandler{Client: client}
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("HereSphere-JSON-Version", "1"))
	r.Post("/", indexHandler(httpHandler.Index))
	r.Post("/{videoId}", videoDataHandler(httpHandler.VideoData))
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
