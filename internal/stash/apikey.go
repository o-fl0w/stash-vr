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
// is configured. Use this for URLs that will be sent to clients.
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

// InternalUrl appends the configured Stash API key to the URL as a query
// parameter WITHOUT rebasing to STASH_BASE_URL. Use this for server-side
// fetches made by stash-vr itself (e.g. fetching images to composite),
// where the public-facing URL may not be reachable from inside the container.
func InternalUrl(rawUrl string) string {
	apiKey := config.Application().StashApiKey
	if apiKey == "" || strings.Contains(rawUrl, "apikey") {
		return rawUrl
	}
	if strings.Contains(rawUrl, "?") {
		return rawUrl + "&apikey=" + apiKey
	}

	return rawUrl + "?apikey=" + apiKey
}
