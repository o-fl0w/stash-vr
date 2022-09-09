package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"net/http"
	"stash-vr/internal/config"
	"stash-vr/internal/logger"
	"stash-vr/internal/router"
)

//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

const listenAddress = ":9666"

func main() {
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisableCapacities = true
	if err := run(); err != nil {
		logger.Get().Warn().Err(err).Msg("EXIT with ERROR")
	} else {
		logger.Get().Info().Msg("EXIT without error")
	}
}

func run() error {
	logger.Get().Info().Str("config", fmt.Sprintf("%+v", config.Get())).Send()

	r := router.Build()

	logger.Get().Info().Msg(fmt.Sprintf("Server listening on %s", listenAddress))
	err := http.ListenAndServe(listenAddress, r)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
