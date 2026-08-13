package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"

	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/application"
	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/config"
	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/infrastructure/providers"
)

func main() {
	cfg := config.Load()

	// Логгер
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Logger()

	// Подключение к RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer conn.Close()

	// Создаём exchange и очереди (если ещё не созданы)
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open channel")
	}
	if err := rabbitmq.Setup(ch); err != nil {
		logger.Fatal().Err(err).Msg("failed to setup rabbitmq")
	}
	ch.Close()

	// Инициализируем провайдеров
	emailProvider := providers.NewEmailProvider(&logger)
	smsProvider := providers.NewSMSProvider(&logger)
	pushProvider := providers.NewPushProvider(&logger)

	// Обработчик
	handler := application.NewNotificationHandler(emailProvider, smsProvider, pushProvider, &logger)

	// Создаём консьюмеров для каждой очереди
	consumers := []struct {
		queue   string
		handler func(ctx context.Context, body []byte) error
	}{
		{rabbitmq.QueueEmail, handler.HandleEmail},
		{rabbitmq.QueueSMS, handler.HandleSMS},
		{rabbitmq.QueuePush, handler.HandlePush},
	}

	var consumerInstances []*rabbitmq.Consumer
	for _, c := range consumers {
		consumer, err := rabbitmq.NewConsumer(conn, c.queue, c.handler, &logger)
		if err != nil {
			logger.Fatal().Err(err).Str("queue", c.queue).Msg("failed to create consumer")
		}
		consumerInstances = append(consumerInstances, consumer)
	}

	// Запускаем консьюмеров в отдельных горутинах
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, c := range consumerInstances {
		go func(cons *rabbitmq.Consumer) {
			if err := cons.Start(ctx); err != nil {
				logger.Error().Err(err).Msg("consumer stopped with error")
			}
		}(c)
	}

	logger.Info().Msg("notification service started, waiting for messages...")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down gracefully...")
	cancel()

	// Останавливаем консьюмеров и закрываем соединения
	for _, c := range consumerInstances {
		c.Stop()
		c.Close()
	}
}
