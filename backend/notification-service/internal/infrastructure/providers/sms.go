package providers

import (
	"context"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/domain"
	"github.com/rs/zerolog"
)

type SMSProvider struct {
	logger *zerolog.Logger
}

func NewSMSProvider(logger *zerolog.Logger) *SMSProvider {
	return &SMSProvider{logger: logger}
}

func (p *SMSProvider) Send(ctx context.Context, msg domain.Notification) error {
	p.logger.Info().
		Str("user_id", msg.UserID).
		Str("type", msg.Type).
		Float64("amount", msg.Amount).
		Str("currency", msg.Currency).
		Msg("sending SMS notification")
	time.Sleep(50 * time.Millisecond)
	return nil
}
