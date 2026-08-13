package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "notifications"
	ExchangeType = "topic"

	QueueEmail = "notification.email"
	QueueSMS   = "notification.sms"
	QueuePush  = "notification.push"

	RoutingKeyEmail = "notification.email"
	RoutingKeySMS   = "notification.sms"
	RoutingKeyPush  = "notification.push"

	DLQPrefix = "dlq."
)

// Setup создаёт exchange, очереди и DLQ
func Setup(ch *amqp.Channel) error {
	// 1. Обменник
	err := ch.ExchangeDeclare(
		ExchangeName,
		ExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Очереди и их DLQ
	queues := []string{QueueEmail, QueueSMS, QueuePush}
	for _, q := range queues {
		// Основная очередь с DLQ
		_, err := ch.QueueDeclare(
			q,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			amqp.Table{
				"x-dead-letter-exchange":    ExchangeName,
				"x-dead-letter-routing-key": DLQPrefix + q,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q, err)
		}

		// DLQ очередь (без дополнительных параметров)
		dlqName := DLQPrefix + q
		_, err = ch.QueueDeclare(
			dlqName,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare DLQ %s: %w", dlqName, err)
		}

		// Привязка основной очереди к exchange (routing key = имя очереди)
		err = ch.QueueBind(
			q,
			q, // routing key = notification.email и т.д.
			ExchangeName,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q, err)
		}

		// Привязка DLQ к exchange (чтобы отправлять в DLQ по routing key dlq.notification.email)
		err = ch.QueueBind(
			dlqName,
			DLQPrefix+q,
			ExchangeName,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind DLQ %s: %w", dlqName, err)
		}
	}

	return nil
}
