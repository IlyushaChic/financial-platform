package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/event"
)

// KafkaPublisher реализует интерфейс messaging.EventPublisher для Kafka
type KafkaPublisher struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaPublisher создаёт новый экземпляр продюсера
func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0                 // можно использовать более новую
	config.Producer.RequiredAcks = sarama.WaitForAll // acks=all
	config.Producer.Idempotent = true                // идемпотентность
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	config.Net.MaxOpenRequests = 1 // необходимо для идемпотентности

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaPublisher{
		producer: producer,
		topic:    topic,
	}, nil
}

// SendMessageWithHeaders отправляет произвольное сообщение с заголовками
func (p *KafkaPublisher) SendMessageWithHeaders(msg *sarama.ProducerMessage) error {
	_, _, err := p.producer.SendMessage(msg)
	return err
}

// Publish отправляет событие в Kafka
func (p *KafkaPublisher) Publish(ctx context.Context, evt interface{}) error {
	eventData, ok := evt.(event.TransactionCompletedEvent)
	if !ok {
		return fmt.Errorf("unsupported event type: %T", evt)
	}

	payload, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(eventData.TransactionID), // партиционирование по ID
		Value: sarama.ByteEncoder(payload),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event-type"), Value: []byte("TransactionCompleted")},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// можно добавить логирование успеха
	_ = partition
	_ = offset
	return nil
}

// Close закрывает продюсера
func (p *KafkaPublisher) Close() error {
	return p.producer.Close()
}

// PublishDLQ отправляет сообщение в топик DLQ
func (p *KafkaPublisher) PublishDLQ(ctx context.Context, originalMsg []byte, reason string) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic + "_dlq", // например, "transactions_dlq"
		Value: sarama.ByteEncoder(originalMsg),
		Headers: []sarama.RecordHeader{
			{Key: []byte("error"), Value: []byte(reason)},
			{Key: []byte("timestamp"), Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}
	_, _, err := p.producer.SendMessage(msg)
	return err
}
