//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

package main

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/api/common"
	"stash-vr/internal/application"
	"stash-vr/internal/config"
	_ "stash-vr/internal/logger"
	"stash-vr/internal/server"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
)

const listenAddress = ":9666"

func main() {
	if err := run(); err != nil {
		log.Warn().Err(err).Msg("Application EXIT with ERROR")
	} else {
		log.Info().Msg("Application EXIT without error")
	}
}

func run() error {
	ctx := application.InterruptableContext()

	log.Info().Str("config", fmt.Sprintf("%+v", config.Get().Redacted())).Send()

	stashClient := stash.NewClient(config.Get().StashGraphQLUrl, config.Get().StashApiKey)

	logVersions(ctx, stashClient)

	common.GetIndex(ctx, stashClient)

	err := server.Listen(ctx, listenAddress, stashClient)
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}

func logVersions(ctx context.Context, client graphql.Client) {
	log.Info().Str("Stash-VR version", application.BuildVersion).Send()

	if version, err := gql.Version(ctx, client); err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve stash version")
	} else {
		log.Info().Str("Stash version", version.Version.Version).Send()
	}
}
