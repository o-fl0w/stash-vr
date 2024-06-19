package heresphere

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"net/url"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := httpHandler{StashClient: client}
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("HereSphere-JSON-Version", "1"))
	r.Post("/", internal.LogRoute("index", httpHandler.indexHandler))
	r.Post("/scan", internal.LogRoute("scan", httpHandler.scanHandler))
	if config.Get().HspDir != "" {
		r.Get("/{videoId}/hsp", internal.LogRoute("hsp", httpHandler.videoHspHandler(config.Get().HspDir)).ServeHTTP)
	}
	r.Handle("/{videoId}", internal.LogRoute("videoData", internal.LogVideoId(httpHandler.videoDataHandler)))
	r.Post("/events", internal.LogRoute("events", httpHandler.eventsHandler))
	return r
}

func getVideoDataUrl(baseUrl string, id string) string {
	return baseUrl + "/heresphere/" + url.QueryEscape(id)
}

func getHspDataUrl(baseUrl string, id string) string {
	return baseUrl + "/heresphere/" + url.QueryEscape(id) + "/hsp"
}
