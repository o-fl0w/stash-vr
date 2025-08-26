package stash

import (
	"stash-vr/internal/config"
	"strings"
)

func ApiKeyed(url string) string {
	apiKey := config.Application().StashApiKey
	if apiKey == "" || strings.Contains(url, "apikey") {
		return url
	}
	if strings.Contains(url, "?") {
		return url + "&apikey=" + apiKey
	}

	return url + "?apikey=" + apiKey
}
