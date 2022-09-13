package stash

import (
	"github.com/Khan/genqlient/graphql"
	"net/http"
)

type authTransport struct {
	key string
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("ApiKey", t.key)
	return http.DefaultTransport.RoundTrip(req)
}

func NewClient(graphqlUrl string, apiKey string) graphql.Client {
	htc := http.Client{}
	if apiKey != "" {
		htc.Transport = authTransport{key: apiKey}
	}
	return graphql.NewClient(graphqlUrl, &htc)
}
