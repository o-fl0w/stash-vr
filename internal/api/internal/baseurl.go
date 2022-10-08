package internal

import (
	"fmt"
	"net/http"
	"stash-vr/internal/util"
)

func GetBaseUrl(req *http.Request) string {
	scheme := util.GetScheme(req)
	return fmt.Sprintf("%s://%s", scheme, req.Host)
}
