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
	"stash-vr/internal/util"
	"stash-vr/internal/web"
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

	router.Get("/", redirector(client))
	router.Get("/*", web.ServeStatic())

	return router
}

func redirector(client graphql.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")

		if strings.Contains(userAgent, "HereSphere") {
			log.Ctx(r.Context()).Trace().Msg("Redirecting to /heresphere")
			http.Redirect(w, r, "/heresphere", 307)
			return
		}

		web.ServeIndex(client).ServeHTTP(w, r)
		return
	}
}

func logDecorator(next http.Handler, mod string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := log.With().Str("mod", mod).Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := util.GetScheme()
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
