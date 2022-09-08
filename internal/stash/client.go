package stash

import (
	"github.com/Khan/genqlient/graphql"
	"net/http"
	"stash-vr/internal/config"
)

type authTransport struct {
	key string
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("ApiKey", t.key)
	return http.DefaultTransport.RoundTrip(req)
}

func NewClient() graphql.Client {
	htc := http.Client{}
	apiKey := config.Get().StashApiKey
	if apiKey != "" {
		htc.Transport = authTransport{key: apiKey}
	}
	return graphql.NewClient(config.Get().StashGraphQLUrl, &htc)
}
