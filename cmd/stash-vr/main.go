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
	logger.Log.Info().Str("config", fmt.Sprintf("%#v", cfg)).Send()

	r := router.Build(cfg)

	logger.Log.Info().Str("address", listenAddress).Msg("Starting server...")
	err := http.ListenAndServe(listenAddress, r)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
