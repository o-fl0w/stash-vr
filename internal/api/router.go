package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/deovr"
	"stash-vr/internal/api/heatmap"
	"stash-vr/internal/api/heresphere"
	"stash-vr/internal/api/web"
	"stash-vr/internal/config"
	"stash-vr/internal/library"
	"stash-vr/internal/util"
	"strings"
	"time"
)

func Router(libraryService *library.Service) *chi.Mux {
	router := chi.NewRouter()

	router.Use(requestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5, "application/json"))

	//router.Mount("/debug", middleware.Profiler())

	router.Mount("/heresphere", logMod("heresphere", heresphere.Router(libraryService)))
	router.Mount("/deovr", logMod("deovr", deovr.Router(libraryService)))

	router.Post("/filters", logMod("filters", web.FiltersUpdateHandler()).ServeHTTP)

	router.Get("/", rootHandler(libraryService))
	router.Get("/*", logMod("static", staticHandler()).ServeHTTP)

	router.Get("/cover/{videoId}", logMod("heatmap", heatmap.CoverHandler(libraryService)).ServeHTTP)

	return router
}

func rootHandler(libraryService *library.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")

		if strings.Contains(userAgent, "HereSphere") {
			log.Ctx(r.Context()).Trace().Msg("Redirecting to /heresphere")
			http.Redirect(w, r, "/heresphere", 307)
		} else {
			logMod("web", web.IndexHandler(libraryService)).ServeHTTP(w, r)
		}
	}
}

func logMod(value string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := log.Ctx(r.Context()).With().Str("mod", value).Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := util.GetScheme(r)
		url := scheme + "://" + config.Redacted(r.Host) + r.RequestURI

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

func staticHandler() http.HandlerFunc {
	filesDir := http.Dir("./web/static")
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rCtx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		fs.ServeHTTP(w, r)
	}
}
