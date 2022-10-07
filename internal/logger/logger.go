package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"stash-vr/internal/config"
	"stash-vr/internal/util"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "Jan 02, 15:04:05",
	}).Level(zerolog.TraceLevel).With().Caller().Logger()

	level, err := zerolog.ParseLevel(config.Get().LogLevel)
	if err != nil {
		panic(fmt.Sprintf("error parsing log level: %v", err))
	}

	log.Logger = log.Logger.Level(level)

	zerolog.DefaultContextLogger = util.Ptr(WithMod("default"))
}

func WithMod(mod string) zerolog.Logger {
	return log.With().Str("mod", mod).Logger()
}
