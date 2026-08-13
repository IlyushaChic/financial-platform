package providers

import (
	"context"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/domain"
	"github.com/rs/zerolog"
)

type PushProvider struct {
	logger *zerolog.Logger
}

func NewPushProvider(logger *zerolog.Logger) *PushProvider {
	return &PushProvider{logger: logger}
}

func (p *PushProvider) Send(ctx context.Context, msg domain.Notification) error {
	p.logger.Info().
		Str("user_id", msg.UserID).
		Str("type", msg.Type).
		Float64("amount", msg.Amount).
		Str("currency", msg.Currency).
		Msg("sending PUSH notification")
	time.Sleep(50 * time.Millisecond)
	return nil
}
