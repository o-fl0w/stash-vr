package router

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/deovr"
	"stash-vr/internal/api/heresphere"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"strings"
	"time"
)

func Build() *chi.Mux {
	gqlClient := stash.NewClient()

	router := chi.NewRouter()

	router.Use(requestLogger)

	router.Mount("/heresphere", heresphere.Router(gqlClient))
	router.Mount("/deovr", deovr.Router(gqlClient))

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
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s%s", scheme, config.Redacted(r.Host), r.RequestURI)

		baseLogger := log.With().
			Str("method", r.Method).
			Str("url", url).Logger()

		baseLogger.Trace().
			Str("proto", r.Proto).
			Str("user_agent", r.UserAgent()).
			Msg("Incoming request")

		start := time.Now()
		next.ServeHTTP(w, r)

		baseLogger.Trace().
			Dur("ms", time.Since(start)).
			Msg("Request handled")
	})
}
