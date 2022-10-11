package deovr

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"net/http"
	"stash-vr/internal/api/internal"
	"strings"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := httpHandler{Client: client}
	r := chi.NewRouter()

	r.Get("/", internal.LogRoute("index", httpHandler.indexHandler))
	r.Get("/{videoId}", internal.LogRoute("videoData", internal.LogVideoId(httpHandler.videoDataHandler)))
	return r
}

func getVideoDataUrl(baseUrl string, id string) string {
	path := "/deovr/"
	sb := strings.Builder{}
	sb.Grow(len(baseUrl) + len(path) + len(id))
	sb.WriteString(baseUrl)
	sb.WriteString(path)
	sb.WriteString(id)
	return sb.String()
}
