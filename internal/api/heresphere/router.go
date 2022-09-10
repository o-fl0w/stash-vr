package heresphere

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func Router(client graphql.Client) http.Handler {
	httpHandler := HttpHandler{Client: client}
	r := chi.NewRouter()
	r.Use(logContext)
	r.Use(addHSHeader)
	r.Post("/", httpHandler.Index)
	r.Post("/{videoId}", httpHandler.VideoData)
	return r
}

func logContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := log.With().Str("module", "deovr").Logger().WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addHSHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("HereSphere-JSON-Version", "1")
		next.ServeHTTP(w, r)
	})
}
