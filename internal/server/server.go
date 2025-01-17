package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"stash-vr/internal/ivdb"
	"stash-vr/internal/router"
	"stash-vr/internal/stimhub"
	"time"
)

func Listen(ctx context.Context, listenAddress string, stashClient graphql.Client, stimhubClient *stimhub.Client, ivdbClient *ivdb.Client) error {
	server := http.Server{
		Addr:    listenAddress,
		Handler: router.Build(stashClient, stimhubClient, ivdbClient),
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Ctx(ctx).Info().Msg(fmt.Sprintf("Server listening at %s", listenAddress))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			<-ctxShutdown.Done()
			if errors.Is(ctxShutdown.Err(), context.DeadlineExceeded) {
				log.Ctx(ctx).Warn().Err(ctxShutdown.Err()).Msg("Shutdown timed out")
			}
		}()

		if err := server.Shutdown(ctxShutdown); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Server shutdown error")
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	log.Ctx(ctx).Debug().Msg("Server stopped without error")
	return nil
}
