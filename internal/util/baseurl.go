package util

import (
	"fmt"
	"net/http"
)

func GetBaseUrl(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, req.Host)
}
