package interrupt

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func Context() context.Context {
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
