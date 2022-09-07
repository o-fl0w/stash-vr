package main

import (
	"fmt"
	"net/http"
	"stash-vr/internal/config"
	"stash-vr/internal/logger"
	"stash-vr/internal/router"
)

//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

const listenAddress = ":9666"

func main() {
	cfg := config.Load()

	if err := run(cfg); err != nil {
		logger.Log.Warn().Err(err).Msg("EXIT with ERROR")
	} else {
		logger.Log.Info().Msg("EXIT without error")
	}
}

func run(cfg config.Application) error {
	logger.Log.Info().Msg(fmt.Sprintf("%#v", cfg))

	r := router.Build(cfg)
	//r := LogRouter{router.Build(cfg)}

	err := http.ListenAndServe(listenAddress, r)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

type LogRouter struct {
	h http.Handler
}

func (lr LogRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().
		Str("method", r.Method).
		Str("uri", r.URL.String()).
		Str("user-agent", r.Header.Get("User-Agent")).
		Str("remote addr", r.RemoteAddr).Send()
	lr.h.ServeHTTP(w, r)
}
