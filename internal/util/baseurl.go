package util

import (
	"fmt"
	"net/http"
	"stash-vr/internal/config"
)

func GetScheme() string {
	if config.Get().IsHTTPS {
		return "https"
	}
	return "http"
}

func GetBaseUrl(req *http.Request) string {
	scheme := GetScheme()
	return fmt.Sprintf("%s://%s", scheme, req.Host)
}
