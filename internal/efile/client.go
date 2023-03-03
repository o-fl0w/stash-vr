package efile

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type eFileIndex struct {
	EFileNames []string `json:"eFileNames"`
}

func GetEFileNames(eventServerUrl string) ([]string, error) {
	resp, err := http.Get(eventServerUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to GET EFiles from eventServer: %w", err)
	}
	var efi eFileIndex
	err = json.NewDecoder(resp.Body).Decode(&efi)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response from eventServer: %w", err)
	}
	resp.Body.Close()
	return efi.EFileNames, err
}
