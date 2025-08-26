package util

import (
	"net/http"
	"stash-vr/internal/config"
)

func GetScheme(req *http.Request) string {
	if req.URL.Scheme == "https" || req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" || config.Application().ForceHTTPS {
		return "https"
	}
	return "http"
}
