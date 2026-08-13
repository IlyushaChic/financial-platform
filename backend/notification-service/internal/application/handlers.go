package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/domain"
	"github.com/IlyushaChic/financial-platform/backend/notification-service/internal/infrastructure/providers"
	"github.com/rs/zerolog"
)

type NotificationHandler struct {
	emailProvider *providers.EmailProvider
	smsProvider   *providers.SMSProvider
	pushProvider  *providers.PushProvider
	logger        *zerolog.Logger
}

func NewNotificationHandler(
	email *providers.EmailProvider,
	sms *providers.SMSProvider,
	push *providers.PushProvider,
	logger *zerolog.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		emailProvider: email,
		smsProvider:   sms,
		pushProvider:  push,
		logger:        logger,
	}
}

func (h *NotificationHandler) HandleEmail(ctx context.Context, body []byte) error {
	var msg domain.Notification
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal email: %w", err)
	}
	h.logger.Info().Interface("msg", msg).Msg("handling email")
	return h.emailProvider.Send(ctx, msg)
}

func (h *NotificationHandler) HandleSMS(ctx context.Context, body []byte) error {
	var msg domain.Notification
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal sms: %w", err)
	}
	h.logger.Info().Interface("msg", msg).Msg("handling sms")
	return h.smsProvider.Send(ctx, msg)
}

func (h *NotificationHandler) HandlePush(ctx context.Context, body []byte) error {
	var msg domain.Notification
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal push: %w", err)
	}
	h.logger.Info().Interface("msg", msg).Msg("handling push")
	return h.pushProvider.Send(ctx, msg)
}
