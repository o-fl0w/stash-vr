package util

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func NewJsonEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

func AsJsonStr(obj interface{}) string {
	sb := new(strings.Builder)
	enc := json.NewEncoder(sb)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")
	_ = enc.Encode(obj)
	compacted := bytes.Buffer{}
	_ = json.Compact(&compacted, []byte(sb.String()))
	return compacted.String()
}
