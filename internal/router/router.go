package router

import (
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/deovr"
	"stash-vr/internal/api/heresphere"
	"stash-vr/internal/config"
	"strings"
	"time"
)

func Build(client graphql.Client) *chi.Mux {
	router := chi.NewRouter()

	router.Use(requestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5, "application/json"))

	//router.Mount("/debug", middleware.Profiler())

	router.Mount("/heresphere", logDecorator(heresphere.Router(client), "heresphere"))
	router.Mount("/deovr", logDecorator(deovr.Router(client), "deovr"))

	router.Get("/", redirector)

	return router
}

func redirector(w http.ResponseWriter, req *http.Request) {
	userAgent := req.Header.Get("User-Agent")

	if strings.Contains(userAgent, "HereSphere") {
		log.Ctx(req.Context()).Trace().Msg("Redirecting to /heresphere")
		http.Redirect(w, req, "/heresphere", 307)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func logDecorator(next http.Handler, mod string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := log.With().Str("mod", mod).Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s%s", scheme, config.Redacted(r.Host), r.RequestURI)

		baseLogger := log.Ctx(r.Context()).With().
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
