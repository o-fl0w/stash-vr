package util

import (
	"encoding/json"
	"io"
)

func NewJsonEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}
