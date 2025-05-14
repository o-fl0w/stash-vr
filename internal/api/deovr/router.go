package deovr

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/library"
)

func Router(libraryService *library.Service) http.Handler {
	httpHandler := httpHandler{libraryService}
	r := chi.NewRouter()

	r.Get("/", internal.LogRoute("index", httpHandler.indexHandler))
	r.Get("/{videoId}", internal.LogRoute("videoData", internal.LogVideoId(httpHandler.videoDataHandler)))
	return r
}

func getVideoDataUrl(baseUrl string, id string) string {
	return baseUrl + "/deovr/" + id
}
