package heresphere

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"net/url"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/library"
	"stash-vr/internal/static"
)

func Router(libraryService *library.Service) http.Handler {
	httpHandler := httpHandler{libraryService: libraryService}
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("HereSphere-JSON-Version", "1"))
	r.Post("/", internal.LogRoute("index", httpHandler.indexHandler))
	r.Get("/", internal.LogRoute("static", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFileFS(writer, request, static.Fs, "loading.html")
	}))

	r.Post("/scan", internal.LogRoute("scan", httpHandler.scanHandler))
	r.Post("/auth", http.NotFound)
	r.Handle("/{videoId}", internal.LogRoute("videoData", internal.LogVideoId(httpHandler.videoDataHandler)))
	r.Post("/events/{videoId}", internal.LogRoute("events", httpHandler.eventsHandler))
	return r
}

func getVideoDataUrl(baseUrl string, id string) string {
	return baseUrl + "/heresphere/" + url.QueryEscape(id)
}

func getEventsUrl(baseUrl string, id string) string {
	return baseUrl + "/heresphere/events/" + url.QueryEscape(id)
}
