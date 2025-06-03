package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
)

func newJsonEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

func WriteJson(ctx context.Context, w http.ResponseWriter, data any) error {
	buf := bytes.Buffer{}
	err := newJsonEncoder(&buf).Encode(data)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	log.Ctx(ctx).Trace().Str("length", byteCountDecimal(buf.Len())).Msg("About to write response")

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.Itoa(buf.Len()))

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func byteCountDecimal(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
