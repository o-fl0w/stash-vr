package stimhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type StimSceneList struct {
	StimScenes []StimScene `json:"stimScenes"`
}
type StimScene struct {
	AudioCrc32 string        `json:"audioCrc32"`
	SceneId    string        `json:"sceneId"`
	Title      string        `json:"title,omitempty"`
	Artist     string        `json:"artist,omitempty"`
	DateAdded  time.Time     `json:"dateAdded"`
	FileNames  []string      `json:"fileNames"`
	Duration   time.Duration `json:"duration"`
}

type Client struct {
	Endpoint string
}

func (c Client) EventServerUrl() string {
	u, _ := url.JoinPath(c.Endpoint, "event")
	return u
}

func (c Client) ThumbnailUrl(audioCrc32 string) string {
	u, _ := url.JoinPath(c.Endpoint, "cover", audioCrc32)
	return u
}

func StimScenes(ctx context.Context, client Client) ([]StimScene, error) {
	u, _ := url.JoinPath(client.Endpoint, "stimscenes")
	request, err := http.NewRequestWithContext(ctx, "get", u, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	var list StimSceneList

	err = json.NewDecoder(response.Body).Decode(&list)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	response.Body.Close()
	return list.StimScenes, nil
}
