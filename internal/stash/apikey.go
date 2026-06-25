package stash

import (
	"net/url"
	"stash-vr/internal/config"
	"strings"
)

// RebaseUrl replaces the scheme, host, and port of a Stash-origin URL with
// the base URL configured via STASH_BASE_URL. This allows the server to
// connect to Stash internally while directing clients to the public-facing
// reverse proxy address. If STASH_BASE_URL is not set the original URL is
// returned unchanged.
func RebaseUrl(rawUrl string) string {
	stashBaseUrl := config.Application().StashBaseUrl
	if stashBaseUrl == "" {
		return rawUrl
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return rawUrl
	}

	base, err := url.Parse(stashBaseUrl)
	if err != nil {
		return rawUrl
	}

	parsed.Scheme = base.Scheme
	parsed.Host = base.Host

	return parsed.String()
}

// ApiKeyed appends the configured Stash API key to the URL as a query
// parameter. The URL origin is first rewritten via RebaseUrl if STASH_BASE_URL
// is configured.
func ApiKeyed(rawUrl string) string {
	u := RebaseUrl(rawUrl)

	apiKey := config.Application().StashApiKey
	if apiKey == "" || strings.Contains(u, "apikey") {
		return u
	}
	if strings.Contains(u, "?") {
		return u + "&apikey=" + apiKey
	}

	return u + "?apikey=" + apiKey
}
