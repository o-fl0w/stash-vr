package main

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"net/http"
	"stash-vr/internal/api/common"
	"stash-vr/internal/config"
	_ "stash-vr/internal/logger"
	"stash-vr/internal/router"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

const listenAddress = ":9666"

var BuildVersion = "DEV"

func main() {
	if err := run(); err != nil {
		log.Warn().Err(err).Msg("EXIT with ERROR")
	} else {
		log.Info().Msg("EXIT without error")
	}
}

func run() error {
	ctx := context.Background()

	log.Info().Str("config", fmt.Sprintf("%+v", config.Get().Redacted())).Send()

	stashClient := stash.NewClient(config.Get().StashGraphQLUrl, config.Get().StashApiKey)

	logVersions(ctx, stashClient)

	common.GetIndex(ctx, stashClient)

	server := http.Server{
		Addr:    listenAddress,
		Handler: router.Build(stashClient),
	}

	log.Info().Msg(fmt.Sprintf("Server listening on %s", listenAddress))
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}

func logVersions(ctx context.Context, client graphql.Client) {
	log.Info().Str("stash-vr version", BuildVersion).Send()

	if version, err := gql.Version(ctx, client); err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve stash version")
	} else {
		log.Info().Str("stash version", version.Version.Version).Send()
	}
}
