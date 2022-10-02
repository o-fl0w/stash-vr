package web

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

func ServeStatic() http.HandlerFunc {
	filesDir := http.Dir("./web/static")
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rCtx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		fs.ServeHTTP(w, r)
	}
}
