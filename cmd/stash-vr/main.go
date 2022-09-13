package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/config"
	_ "stash-vr/internal/logger"
	"stash-vr/internal/router"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

const listenAddress = ":9666"

func main() {
	if err := run(); err != nil {
		log.Warn().Err(err).Msg("EXIT with ERROR")
	} else {
		log.Info().Msg("EXIT without error")
	}
}

func run() error {
	log.Info().Str("config", fmt.Sprintf("%+v", config.Get().Redacted())).Send()

	stashClient := stash.NewClient(config.Get().StashGraphQLUrl, config.Get().StashApiKey)

	if version, err := gql.Version(context.Background(), stashClient); err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve stash version")
	} else {
		log.Info().Str("stash version", version.Version.Version).Send()
	}

	r := router.Build(stashClient)

	log.Info().Msg(fmt.Sprintf("Server listening on %s", listenAddress))
	err := http.ListenAndServe(listenAddress, r)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}
