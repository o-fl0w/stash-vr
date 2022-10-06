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
	"stash-vr/internal/api/web"
	"stash-vr/internal/config"
	"stash-vr/internal/logger"
	"stash-vr/internal/util"
	"strings"
	"time"
)

func Build(client graphql.Client) *chi.Mux {
	router := chi.NewRouter()

	router.Use(requestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5, "application/json"))

	//router.Mount("/debug", middleware.Profiler())

	router.Mount("/heresphere", logMod("heresphere", heresphere.Router(client)))
	router.Mount("/deovr", logMod("deovr", deovr.Router(client)))

	router.Get("/", rootHandler(client))
	router.Get("/*", logMod("static", staticHandler()).ServeHTTP)

	return router
}

func rootHandler(client graphql.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")

		if strings.Contains(userAgent, "HereSphere") {
			log.Ctx(r.Context()).Trace().Msg("Redirecting to /heresphere")
			http.Redirect(w, r, "/heresphere", 307)
		} else {
			logMod("web", web.IndexHandler(client)).ServeHTTP(w, r)
		}
	}
}

func logMod(value string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithMod(value).WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := util.GetScheme(r)
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
