package logger

import (
	"bytes"
	"encoding/json"
	"strings"
)

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
