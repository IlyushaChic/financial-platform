package providers

import (
	"context"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/domain"
	"github.com/rs/zerolog"
)

type EmailProvider struct {
	logger *zerolog.Logger
}

func NewEmailProvider(logger *zerolog.Logger) *EmailProvider {
	return &EmailProvider{logger: logger}
}

func (p *EmailProvider) Send(ctx context.Context, msg domain.Notification) error {
	p.logger.Info().Str("user_id", msg.UserID).Interface("data", msg.Data).Msg("sending email")
	time.Sleep(50 * time.Millisecond)
	return nil
}
