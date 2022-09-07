package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

var Log = log.Output(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: "Jan 02, 15:04:05",
})

func WithModule(s string) *zerolog.Logger {
	l := Log.With().Str("module", s).Logger()
	return &l
}
