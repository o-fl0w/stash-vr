package efile

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ListResponse struct {
	EScenes []EScene `json:"links"`
}

type EScene struct {
	SceneId   string        `json:"sceneId"`
	Title     string        `json:"title,omitempty"`
	Artist    string        `json:"artist,omitempty"`
	Duration  time.Duration `json:"duration"`
	AddedTime time.Time     `json:"addedTime"`
	Oshash    string        `json:"oshash"`
	FileNames []string      `json:"fileNames"`
}

func GetList(eSceneServer string) ([]EScene, error) {
	u, _ := url.JoinPath(eSceneServer, "/list")
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("failed to GET /list: %w", err)
	}
	var list ListResponse
	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	resp.Body.Close()
	return list.EScenes, nil
}

func GetEScene(eSceneServer string, oshash string, sceneId string) (EScene, error) {
	u, _ := url.Parse(eSceneServer + "/escene")
	q, _ := url.ParseQuery(u.RawQuery)
	q.Add("oshash", oshash)
	q.Add("sceneId", sceneId)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return EScene{}, fmt.Errorf("failed to GET /escene: %w", err)
	}
	var eScene EScene
	err = json.NewDecoder(resp.Body).Decode(&eScene)
	if err != nil {
		return EScene{}, fmt.Errorf("failed to parse response: %w", err)
	}
	resp.Body.Close()
	return eScene, nil
}
