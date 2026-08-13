package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/event"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/persistence/clickhouse"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// EventConsumer обрабатывает события из Kafka
type EventConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	dlqTopic      string
	analyticsRepo *clickhouse.AnalyticsRepo
	redisClient   *redis.Client
	logger        *zerolog.Logger
	publisher     *KafkaPublisher
}

// NewEventConsumer создаёт новый консьюмер
func NewEventConsumer(
	brokers []string,
	groupID string,
	topic string,
	dlqTopic string,
	analyticsRepo *clickhouse.AnalyticsRepo,
	redisClient *redis.Client,
	logger *zerolog.Logger,
	publisher *KafkaPublisher,
) (*EventConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &EventConsumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		dlqTopic:      dlqTopic,
		analyticsRepo: analyticsRepo,
		redisClient:   redisClient,
		logger:        logger,
		publisher:     publisher,
	}, nil
}

// Start запускает консьюмера (блокирующий вызов)
func (c *EventConsumer) Start(ctx context.Context) error {
	handler := &consumerHandler{
		analyticsRepo: c.analyticsRepo,
		redisClient:   c.redisClient,
		logger:        c.logger,
		publisher:     c.publisher,
		dlqTopic:      c.dlqTopic,
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
				c.logger.Error().Err(err).Msg("consumer group error")
				return err
			}
		}
	}
}

// Close закрывает консьюмера
func (c *EventConsumer) Close() error {
	return c.consumerGroup.Close()
}

// ---------- Хендлер ----------
type consumerHandler struct {
	analyticsRepo *clickhouse.AnalyticsRepo
	redisClient   *redis.Client
	logger        *zerolog.Logger
	publisher     *KafkaPublisher
	dlqTopic      string
}

const maxRetries = 3

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info().Msg("consumer handler setup")
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info().Msg("consumer handler cleanup")
	return nil
}

// ConsumeClaim — единственная реализация метода
func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.logger.Debug().Msgf("received message: topic=%s, partition=%d, offset=%d", msg.Topic, msg.Partition, msg.Offset)

		// Получаем текущее количество попыток из заголовков
		retryCount := 0
		for _, header := range msg.Headers {
			if string(header.Key) == "retry_count" {
				if val, err := strconv.Atoi(string(header.Value)); err == nil {
					retryCount = val
				}
				break
			}
		}

		var evt event.TransactionCompletedEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			h.logger.Error().Err(err).Msg("failed to unmarshal event")
			h.sendToDLQ(sess, msg, err.Error())
			sess.MarkMessage(msg, "")
			continue
		}

		// Обрабатываем событие
		if err := h.processEvent(sess.Context(), evt); err != nil {
			h.logger.Error().Err(err).Msg("failed to process event")
			if retryCount < maxRetries {
				newRetryCount := retryCount + 1
				h.retryWithBackoff(sess, msg, newRetryCount)
			} else {
				h.sendToDLQ(sess, msg, err.Error())
			}
			sess.MarkMessage(msg, "")
			continue
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}

// retryWithBackoff отправляет сообщение обратно в топик с обновлённым заголовком retry_count
func (h *consumerHandler) retryWithBackoff(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, retryCount int) {
	newMsg := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Key:   sarama.StringEncoder(string(msg.Key)),
		Value: sarama.ByteEncoder(msg.Value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("retry_count"), Value: []byte(strconv.Itoa(retryCount))},
		},
	}
	// Копируем остальные заголовки (кроме retry_count)
	for _, header := range msg.Headers {
		if string(header.Key) != "retry_count" {
			newMsg.Headers = append(newMsg.Headers, sarama.RecordHeader{
				Key:   header.Key,
				Value: header.Value,
			})
		}
	}

	if err := h.publisher.SendMessageWithHeaders(newMsg); err != nil {
		h.logger.Error().Err(err).Msg("failed to retry message")
	} else {
		h.logger.Info().Int("retry", retryCount).Msg("message re-queued for retry")
	}
}

// sendToDLQ отправляет сообщение в DLQ-топик
func (h *consumerHandler) sendToDLQ(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, reason string) {
	if err := h.publisher.PublishDLQ(sess.Context(), msg.Value, reason); err != nil {
		h.logger.Error().Err(err).Msg("failed to send to DLQ")
	} else {
		h.logger.Info().Msg("message sent to DLQ")
	}
}

// processEvent обрабатывает событие (запись в ClickHouse, обновление Redis)
func (h *consumerHandler) processEvent(ctx context.Context, evt event.TransactionCompletedEvent) error {
	// 1. Запись в ClickHouse
	if err := h.analyticsRepo.InsertTransaction(
		ctx,
		evt.TransactionID,
		evt.FromAccountID,
		evt.ToAccountID,
		evt.Amount,
		evt.Currency,
		"completed",
		evt.CompletedAt,
	); err != nil {
		return fmt.Errorf("clickhouse insert failed: %w", err)
	}

	// 2. Обновление кэша баланса в Redis
	if evt.FromAccountID != "" {
		if err := h.redisClient.IncrByFloat(ctx, "balance:"+evt.FromAccountID, -evt.Amount).Err(); err != nil {
			h.logger.Error().Err(err).Msg("failed to update sender balance in redis")
		}
	}
	if evt.ToAccountID != "" {
		if err := h.redisClient.IncrByFloat(ctx, "balance:"+evt.ToAccountID, evt.Amount).Err(); err != nil {
			h.logger.Error().Err(err).Msg("failed to update receiver balance in redis")
		}
	}

	// 3. Агрегаты по валютам
	date := evt.CompletedAt.Truncate(24 * time.Hour)
	if err := h.analyticsRepo.InsertDailyCurrencyAggregate(ctx, date, evt.Currency, evt.Amount, 1); err != nil {
		h.logger.Error().Err(err).Msg("failed to insert daily currency aggregate")
	}

	return nil
}
