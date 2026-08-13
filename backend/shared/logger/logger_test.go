package logger

import (
	"testing"
)

func TestLoggerJSON(t *testing.T) {
	cfg := Config{
		Level: "debug",
		JSON:  true,
	}
	log := New(cfg)
	log.Info().Msg("test info message")
	log.Error().Msg("test error message")
}
