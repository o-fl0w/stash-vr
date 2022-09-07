package router

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"stash-vr/internal/config"
	"stash-vr/internal/deovr"
	"stash-vr/internal/heresphere"
	"strings"
)

func Build(cfg config.Application) *httprouter.Router {
	router := httprouter.New()

	hsHttpHandler := heresphere.HttpHandler{Config: cfg}
	router.POST("/heresphere", hsHttpHandler.Index)
	router.POST("/heresphere/:videoId", hsHttpHandler.VideoData)

	dvHttpHandler := deovr.HttpHandler{Config: cfg}
	router.GET("/deovr", dvHttpHandler.Index)
	router.GET("/deovr/:videoId", dvHttpHandler.VideoData)

	router.GET("/", redirector)

	return router
}

func redirector(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
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
