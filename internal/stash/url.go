package stash

import (
	_url "net/url"
	"stash-vr/internal/config"
)

func ApiKeyed(url string) string {
	u, err := _url.Parse(url)
	if err != nil {
		return ""
	}
	values := u.Query()
	if values.Has("apikey") {
		return url
	}
	values.Set("apikey", config.Get().StashApiKey)
	u.RawQuery = values.Encode()
	s := u.String()
	return s
}
