package heresphere

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"stash-vr/internal/api/heresphere/internal"
	internal2 "stash-vr/internal/api/internal"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := internal.HttpHandler{Client: client}
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("HereSphere-JSON-Version", "1"))
	r.Post("/", internal2.LogRoute("index", httpHandler.Index))
	r.Post("/scan", internal2.LogRoute("scan", httpHandler.Scan))
	r.Post("/{videoId}", internal2.LogRoute("videoData", internal2.LogVideoId(httpHandler.VideoData)))
	return r
}
