package router

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"stash-vr/internal/deovr"
	"stash-vr/internal/heresphere"
	"stash-vr/internal/logger"
	"stash-vr/internal/stash"
	"stash-vr/internal/util"
	"strings"
	"time"
)

func Build() *chi.Mux {
	gqlClient := stash.NewClient()

	router := chi.NewRouter()

	router.Use(requestLogger)

	hsHttpHandler := heresphere.HttpHandler{Client: gqlClient}
	router.Post("/heresphere", hsHttpHandler.Index)
	router.Post("/heresphere/{videoId}", hsHttpHandler.VideoData)

	dvHttpHandler := deovr.HttpHandler{Client: gqlClient}
	router.Get("/deovr", dvHttpHandler.Index)
	router.Get("/deovr/{videoId}", dvHttpHandler.VideoData)

	router.Get("/", redirector)

	return router
}

func redirector(w http.ResponseWriter, req *http.Request) {
	userAgent := req.Header.Get("User-Agent")

	if strings.Contains(userAgent, "HereSphere") {
		http.Redirect(w, req, "/heresphere", 307)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	//else if strings.Contains(userAgent, "Deo VR") {
	//	http.Redirect(w, req, "/deovr", 307)
	//}
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l := logger.Get()

		next.ServeHTTP(w, r)

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s%s", scheme, util.Redacted(r.Host), r.RequestURI)

		l.
			Trace().
			Str("method", r.Method).
			Str("url", url).
			Str("proto", r.Proto).
			Str("user_agent", r.UserAgent()).
			Dur("ms", time.Since(start)).
			Msg("-> request")
	})
}
