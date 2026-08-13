package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	maxRetries = 3
	delayBase  = 2 // секунды, экспоненциальный рост
)

type Consumer struct {
	channel  *amqp.Channel
	queue    string
	handler  func(ctx context.Context, body []byte) error
	logger   *zerolog.Logger
	stopChan chan struct{}
}

func NewConsumer(
	conn *amqp.Connection,
	queue string,
	handler func(ctx context.Context, body []byte) error,
	logger *zerolog.Logger,
) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		channel:  ch,
		queue:    queue,
		handler:  handler,
		logger:   logger,
		stopChan: make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info().Str("queue", c.queue).Msg("consumer started")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Str("queue", c.queue).Msg("consumer stopped by context")
			return nil
		case <-c.stopChan:
			c.logger.Info().Str("queue", c.queue).Msg("consumer stopped by stop signal")
			return nil
		case d, ok := <-deliveries:
			if !ok {
				c.logger.Error().Str("queue", c.queue).Msg("deliveries channel closed")
				return nil
			}

			c.logger.Debug().Str("queue", c.queue).Msgf("received message: %s", d.Body)

			// Извлекаем текущее количество попыток
			retryCount := 0
			for _, header := range d.Headers {
				if key, ok := header.(amqp.Table); ok {
					if val, ok := key["x-retry-count"]; ok {
						if count, ok := val.(int64); ok {
							retryCount = int(count)
						}
					}
				}
			}
			// также можно искать в заголовках сообщения
			if val, ok := d.Headers["x-retry-count"]; ok {
				if count, ok := val.(int64); ok {
					retryCount = int(count)
				}
			}

			// Обрабатываем
			err := c.handler(ctx, d.Body)

			if err != nil {
				c.logger.Error().Err(err).Int("retry", retryCount).Msg("handler error")
				c.handleError(ctx, d, retryCount, err)
				continue
			}

			// Успех – ack
			if err := d.Ack(false); err != nil {
				c.logger.Error().Err(err).Msg("failed to ack")
			}
		}
	}
}

// handleError обрабатывает ошибку: retry или DLQ
func (c *Consumer) handleError(ctx context.Context, d amqp.Delivery, retryCount int, lastErr error) {
	if retryCount >= maxRetries {
		// Отправляем в DLQ
		c.logger.Warn().Int("retry", retryCount).Msg("max retries exceeded, sending to DLQ")
		if err := c.publishToDLQ(d, lastErr); err != nil {
			c.logger.Error().Err(err).Msg("failed to publish to DLQ")
		}
		// Подтверждаем исходное сообщение (оно больше не нужно)
		_ = d.Ack(false)
		return
	}

	// Планируем повторную отправку с экспоненциальной задержкой
	delay := time.Duration(delayBase<<retryCount) * time.Second // 2, 4, 8 секунд
	c.logger.Info().Int("retry", retryCount+1).Dur("delay", delay).Msg("scheduling retry")

	time.AfterFunc(delay, func() {
		if err := c.publishWithRetry(d, retryCount+1); err != nil {
			c.logger.Error().Err(err).Msg("failed to publish retry message")
		}
	})

	// Подтверждаем исходное сообщение (оно будет заменено новым)
	_ = d.Ack(false)
}

// publishWithRetry публикует сообщение обратно в очередь с обновлённым заголовком retry-count
func (c *Consumer) publishWithRetry(original amqp.Delivery, newRetryCount int) error {
	headers := make(amqp.Table)
	for k, v := range original.Headers {
		headers[k] = v
	}
	headers["x-retry-count"] = int64(newRetryCount)

	return c.channel.Publish(
		"", // default exchange
		original.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  original.ContentType,
			Body:         original.Body,
			Headers:      headers,
			DeliveryMode: original.DeliveryMode,
			Timestamp:    time.Now(),
		},
	)
}

// publishToDLQ отправляет сообщение в DLQ
func (c *Consumer) publishToDLQ(original amqp.Delivery, reason error) error {
	headers := make(amqp.Table)
	for k, v := range original.Headers {
		headers[k] = v
	}
	headers["x-error"] = reason.Error()
	headers["x-failed-at"] = time.Now().Format(time.RFC3339)

	dlqRoutingKey := DLQPrefix + original.RoutingKey // например, "dlq.notification.email"

	return c.channel.Publish(
		ExchangeName,
		dlqRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  original.ContentType,
			Body:         original.Body,
			Headers:      headers,
			DeliveryMode: original.DeliveryMode,
			Timestamp:    time.Now(),
		},
	)
}

func (c *Consumer) Stop() {
	close(c.stopChan)
}

func (c *Consumer) Close() error {
	return c.channel.Close()
}
