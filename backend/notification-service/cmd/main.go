package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/config"
	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/infrastructure/messaging/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Caller().Logger()

	cfg := config.Load()
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open channel")
	}
	if err := rabbitmq.Setup(ch); err != nil {
		logger.Fatal().Err(err).Msg("failed to setup rabbitmq")
	}
	ch.Close()

	logger.Info().Msg("Notification service started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("shutting down gracefully...")
}
