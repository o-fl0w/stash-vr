package heresphere

import (
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"stash-vr/internal/api/internal"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := httpHandler{Client: client}
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("HereSphere-JSON-Version", "1"))
	r.Post("/", internal.LogRoute("index", httpHandler.indexHandler))
	r.Post("/scan", internal.LogRoute("scan", httpHandler.scanHandler))
	r.Post("/{videoId}", internal.LogRoute("videoData", internal.LogVideoId(httpHandler.videoDataHandler)))
	return r
}

func getVideoDataUrl(baseUrl string, id string) string {
	return fmt.Sprintf("%s/heresphere/%s", baseUrl, id)
}
