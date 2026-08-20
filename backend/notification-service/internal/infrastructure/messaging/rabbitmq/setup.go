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

func Setup(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		ExchangeName,
		ExchangeType,
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	queues := []string{QueueEmail, QueueSMS, QueuePush}
	for _, q := range queues {
		_, err := ch.QueueDeclare(
			q,     // name
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait  <-- добавлено
			amqp.Table{
				"x-dead-letter-exchange":    ExchangeName,
				"x-dead-letter-routing-key": DLQPrefix + q,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q, err)
		}

		dlqName := DLQPrefix + q
		_, err = ch.QueueDeclare(
			dlqName,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait  <-- добавлено
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare DLQ %s: %w", dlqName, err)
		}

		if err := ch.QueueBind(q, q, ExchangeName, false, nil); err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q, err)
		}
		if err := ch.QueueBind(dlqName, DLQPrefix+q, ExchangeName, false, nil); err != nil {
			return fmt.Errorf("failed to bind DLQ %s: %w", dlqName, err)
		}
	}
	return nil
}
