package internal

import (
	"encoding/json"
	"io"
	"net/http"
)

func UnmarshalBody[T any](req *http.Request) (T, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return *new(T), err
	}

	var structured T
	err = json.Unmarshal(body, &structured)
	if err != nil {
		return *new(T), err
	}
	return structured, nil
}
