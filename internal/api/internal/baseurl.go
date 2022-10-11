package internal

import (
	"net/http"
	"stash-vr/internal/util"
)

func GetBaseUrl(req *http.Request) string {
	scheme := util.GetScheme(req)
	return scheme + "://" + req.Host
}
