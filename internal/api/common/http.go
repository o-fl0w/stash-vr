package common

import (
	"context"
	"fmt"
	"net/http"
	"stash-vr/internal/util"
)

func Write(ctx context.Context, w http.ResponseWriter, data interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	//log.Ctx(ctx).Trace().Msg(fmt.Sprintf("write:\n%s", util.AsJsonStr(data)))
	err := util.NewJsonEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
