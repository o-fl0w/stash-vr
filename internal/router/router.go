package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"stash-vr/internal/deovr"
	"stash-vr/internal/heresphere"
	"stash-vr/internal/stash"
	"strings"
)

func Build() *chi.Mux {
	gqlClient := stash.NewClient()

	router := chi.NewRouter()
	//router.Use(middleware.Logger)

	hsHttpHandler := heresphere.HttpHandler{Client: gqlClient}
	router.Post("/heresphere", hsHttpHandler.Index)
	router.Post("/heresphere/{videoId}", hsHttpHandler.VideoData)

	dvHttpHandler := deovr.HttpHandler{Client: gqlClient}
	router.Get("/deovr", dvHttpHandler.Index)
	router.Get("/deovr/{videoId}", dvHttpHandler.VideoData)

	router.Get("/", redirector)

	return router
}

func redirector(w http.ResponseWriter, req *http.Request) {
	userAgent := req.Header.Get("User-Agent")

	if strings.Contains(userAgent, "HereSphere") {
		http.Redirect(w, req, "/heresphere", 307)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	//else if strings.Contains(userAgent, "Deo VR") {
	//	http.Redirect(w, req, "/deovr", 307)
	//}
}
