package ivdb

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"net/http"
	"net/url"
	"stash-vr/internal/ivdb/api"
	"stash-vr/internal/stash/gql"
	"strings"
)

type authTransport struct {
	key string
}
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Api *api.ClientWithResponses
	Ctx context.Context
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.key))
	return http.DefaultTransport.RoundTrip(req)
}

func NewClient(ctx context.Context, stashClient graphql.Client) (*Client, error) {
	htc := &http.Client{}

	configurationResponse, err := gql.UIConfiguration(ctx, stashClient)
	if err != nil {
		return &Client{}, fmt.Errorf("FindSceneFull: %w", err)
	}
	handyKey := configurationResponse.Configuration.Interface.HandyKey
	if handyKey != "" {
		htc.Transport = authTransport{key: handyKey}
	}
	apiClient, err := api.NewClientWithResponses("https://scripts01.handyfeeling.com/api/script/index/v0", api.WithHTTPClient(htc))
	if err != nil {
		return &Client{}, fmt.Errorf("FindSceneFull: %w", err)
	}
	client := Client{
		Api: apiClient,
		Ctx: ctx,
	}
	return &client, nil
}
func (c Client) getPartnerIDFromURL(inputURL string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	// Get the fragment part after the '#'
	fragment := parsedURL.Fragment
	if fragment == "" {
		return "", fmt.Errorf("no fragment found in the URL")
	}

	// Split the fragment into parts and get the last part
	parts := strings.Split(fragment, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("no valid path in fragment")
	}

	return parts[len(parts)-1], nil
}
func (c Client) GetTokenUrlFromUrl(inputURL string) (string, error) {

	partnerID, err := c.getPartnerIDFromURL(inputURL)
	if err != nil {
		return "", fmt.Errorf("getPartnerIDFromURL: %w", err)
	}

	takeValue := 10 // Replace 10 with the desired limit

	params := api.GetVideoScriptsParams{
		Take: &takeValue,
	}

	videoScriptResponse, err := c.Api.GetVideoScriptsWithResponse(c.Ctx, partnerID, &params)
	if err != nil {
		return "", fmt.Errorf("GetVideoScriptsWithResponse: %w", err)
	}
	if videoScriptResponse.JSON200 != nil {
		scripts := *videoScriptResponse.JSON200 // Dereference once
		if len(scripts) > 0 {
			firstScript := scripts[0]
			tokenUrlResponse, err := c.Api.GetTokenUrlWithResponse(c.Ctx, partnerID, firstScript.ScriptId)
			if err != nil {
				return "", fmt.Errorf("GetTokenUrlWithResponse: %w", err)
			}
			return (*tokenUrlResponse.JSON200).Url, nil

		} else {
			return "", fmt.Errorf("GetTokenUrlFromUr: %w", err)
		}
	}
	return "", fmt.Errorf("GetTokenUrlFromUrl: No Script Found")

}
