package messaging

import "context"

// EventPublisher определяет интерфейс для публикации событий
type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}
