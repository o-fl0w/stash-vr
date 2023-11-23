package internal

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"stash-vr/internal/build"
	"stash-vr/internal/config"
	"stash-vr/internal/logger"
	"stash-vr/internal/sections"
	"stash-vr/internal/server"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/stimhub"
)

func Run(ctx context.Context) error {
	config.Init()
	log.Logger = logger.New(config.Get().LogLevel, config.Get().DisableLogColor)
	zerolog.DefaultContextLogger = &log.Logger

	log.Info().Str("config", fmt.Sprintf("%+v", config.Get().Redacted())).Send()

	stashClient := stash.NewClient(config.Get().StashGraphQLUrl, config.Get().StashApiKey)
	logVersions(ctx, stashClient)

	var stimhubClient *stimhub.Client
	if config.Get().StimhubUrl != "" {
		stimhubClient = &stimhub.Client{Endpoint: config.Get().StimhubUrl}
	}

	sections.Get(ctx, stashClient, stimhubClient)

	err := server.Listen(ctx, config.Get().ListenAddress, stashClient, stimhubClient)
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}

func logVersions(ctx context.Context, client graphql.Client) {
	log.Info().Str("Stash-VR version", build.FullVersion()).Send()

	if version, err := gql.Version(ctx, client); err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve stash version")
	} else {
		log.Info().Str("Stash version", version.Version.Version).Send()
	}
}
