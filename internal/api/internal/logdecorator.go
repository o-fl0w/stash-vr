package internal

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func LogRoute(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := log.Ctx(r.Context()).With().Str("route", route).Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func LogVideoId(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videoId := chi.URLParam(r, "videoId")
		ctx := log.Ctx(r.Context()).With().Str("videoId", videoId).Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
