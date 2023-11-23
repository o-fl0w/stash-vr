package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func New(level string, disableColor bool) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		panic(fmt.Sprintf("error parsing log level: %v", err))
	}

	l := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "Jan 02, 15:04:05",
		NoColor:    disableColor,
	}).With().Str("mod", "default").Logger().Level(lvl) //.With().Caller().Logger()

	return l
}
