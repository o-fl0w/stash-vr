package deovr

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"net/http"
	"stash-vr/internal/api/deovr/internal"
	internal2 "stash-vr/internal/api/internal"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := internal.HttpHandler{Client: client}
	r := chi.NewRouter()

	r.Get("/", internal2.LogRoute("index", httpHandler.Index))
	r.Get("/{videoId}", internal2.LogRoute("videoData", internal2.LogVideoId(httpHandler.VideoData)))
	return r
}
