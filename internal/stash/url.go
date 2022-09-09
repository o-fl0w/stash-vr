package stash

import (
	"stash-vr/internal/config"
	"strings"
)

func ApiKeyed(url string) string {
	if strings.Contains(url, "apikey") {
		return url
	}
	sb := strings.Builder{}
	sb.WriteString(url)
	if strings.Contains(url, "?") {
		sb.WriteString("&")
	} else {
		sb.WriteString("?")
	}
	sb.WriteString("apikey=")
	sb.WriteString(config.Get().StashApiKey)
	s := sb.String()
	return s
}
