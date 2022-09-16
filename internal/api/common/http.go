package common

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/util"
	"strconv"
)

func WriteJson(ctx context.Context, w http.ResponseWriter, data interface{}) error {
	//log.Ctx(ctx).Trace().Msg(fmt.Sprintf("write:\n%s", util.AsJsonStr(data)))

	buf := bytes.Buffer{}
	err := util.NewJsonEncoder(&buf).Encode(data)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	log.Ctx(ctx).Debug().Str("length", byteCountDecimal(buf.Len())).Msg("About to write response")

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
