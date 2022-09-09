package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"stash-vr/internal/config"
	"sync"
)

var once sync.Once

var logger zerolog.Logger

func Get() *zerolog.Logger {
	once.Do(func() {
		level, err := zerolog.ParseLevel(config.Get().LogLevel)
		if err != nil {
			panic(fmt.Sprintf("error parsing log level: %v", err))
		}
		logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "Jan 02, 15:04:05",
		}).Level(level)
	})
	return &logger
}

func WithModule(s string) *zerolog.Logger {
	l := Get().With().Str("module", s).Logger()
	return &l
}
