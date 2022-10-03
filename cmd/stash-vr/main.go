//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

package main

import (
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"stash-vr/internal/api/common"
	"stash-vr/internal/application"
	"stash-vr/internal/config"
	_ "stash-vr/internal/logger"
	"stash-vr/internal/server"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"syscall"
)

const listenAddress = ":9666"

func signalContext() context.Context {
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		s := <-chSignal
		signal.Stop(chSignal)
		close(chSignal)
		log.Ctx(ctx).Info().Stringer("signal", s).Msg("Exit SIGNAL received")
		cancel()
	}()
	return ctx
}

func main() {
	if err := run(); err != nil {
		log.Warn().Err(err).Msg("Application EXIT with ERROR")
	} else {
		log.Info().Msg("Application EXIT without error")
	}
}

func run() error {
	ctx := signalContext()

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
