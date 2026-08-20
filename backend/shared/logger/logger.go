package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	JSON       bool   `env:"LOG_JSON" default:"true"`
	Level      string `env:"LOG_LEVEL" default:"info"`
	TimeFormat string `env:"LOG_TIME_FORMAT" default:"2006-01-02T15:04:05.000Z07:00"`
}

func New(cfg Config) *zerolog.Logger {
	zerolog.TimeFieldFormat = cfg.TimeFormat
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var logger zerolog.Logger
	if cfg.JSON {
		logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Logger()
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().
			Timestamp().
			Caller().
			Logger()
	}

	zerolog.SetGlobalLevel(level)
	return &logger
}

func MustNew(cfg Config) *zerolog.Logger {
	log := New(cfg)
	return log
}
