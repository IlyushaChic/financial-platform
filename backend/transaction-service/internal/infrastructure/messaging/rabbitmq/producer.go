package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "notifications"
)

// Producer отправляет уведомления в RabbitMQ
type Producer struct {
	channel *amqp.Channel
}

// NewProducer создаёт нового продюсера
func NewProducer(conn *amqp.Connection) (*Producer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	return &Producer{channel: ch}, nil
}

// Publish отправляет сообщение в exchange с указанным routing key
func (p *Producer) Publish(ctx context.Context, routingKey string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		ExchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

// Close закрывает канал
func (p *Producer) Close() error {
	return p.channel.Close()
}
