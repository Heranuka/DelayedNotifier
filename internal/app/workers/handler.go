// Package workers provides background notification workers and message handlers.
package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/logger"
)

// Mailer sends a notification message to a destination.
type Mailer interface {
	Send(ctx context.Context, message, destination string) error
}

// NotificationStatusUpdater updates notification delivery status.
type NotificationStatusUpdater interface {
	UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error
	Status(ctx context.Context, noteID uuid.UUID) (string, error)
}

type NotificationHandler struct {
	statusUpdater NotificationStatusUpdater
	email         Mailer
	telegram      Mailer
	log           *logger.ZerologAdapter
}

// NewNotificationHandler creates a RabbitMQ message handler for notifications.
func NewNotificationHandler(
	svc NotificationStatusUpdater,
	email Mailer,
	telegram Mailer,
	log *logger.ZerologAdapter,
) *NotificationHandler {
	return &NotificationHandler{
		statusUpdater: svc,
		email:         email,
		telegram:      telegram,
		log:           log,
	}
}

// Handle processes notification message body.
func (h *NotificationHandler) Handle(ctx context.Context, body []byte) error {
	var notification struct {
		ID          uuid.UUID `json:"id"`
		Message     string    `json:"message"`
		Destination string    `json:"destination"`
		Channel     string    `json:"channel"`
		DataSentAt  time.Time `json:"data_sent_at"` // Matches domain.Notification's tag
	}

	if err := json.Unmarshal(body, &notification); err != nil {
		h.log.Error("failed to unmarshal notification", "err", err, "body", string(body))
		return fmt.Errorf("unmarshal notification: %w", err)
	}

	h.log.Info("[HANDLER] Processing Notification", "id", notification.ID, "channel", notification.Channel)

	// NEW: Check if notification still exists and is not canceled before sending
	status, err := h.statusUpdater.Status(ctx, notification.ID)
	if err != nil {
		h.log.Info("[HANDLER] Notification record not found or error occurred, skipping delivery", "id", notification.ID, "err", err)
		return nil // Ack message as it cannot be processed
	}

	if status == "canceled" {
		h.log.Info("[HANDLER] Notification has been canceled, skipping delivery", "id", notification.ID)
		return nil // Ack message
	}

	if !notification.DataSentAt.IsZero() {
		wait := time.Until(notification.DataSentAt)
		if wait > 0 {
			h.log.Info("[HANDLER] Waiting for scheduled delivery", "id", notification.ID, "wait", wait)
			timer := time.NewTimer(wait)
			defer timer.Stop()

			select {
			case <-timer.C:
				h.log.Info("[HANDLER] Wait finished, sending now", "id", notification.ID)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	var sender Mailer
	switch notification.Channel {
	case "email":
		sender = h.email
	case "telegram":
		sender = h.telegram
	default:
		h.log.Error("unknown channel", "id", notification.ID, "channel", notification.Channel)
		return fmt.Errorf("unknown channel: %s", notification.Channel)
	}

	if sender == nil {
		h.log.Error("sender is nil for channel", "id", notification.ID, "channel", notification.Channel)
		return fmt.Errorf("sender is nil for channel: %s", notification.Channel)
	}

	if err := sender.Send(ctx, notification.Message, notification.Destination); err != nil {
		h.log.Error("[HANDLER] Send failed", "id", notification.ID, "err", err)
		_ = h.statusUpdater.UpdateStatus(ctx, notification.ID, "failed")
		return fmt.Errorf("send failed: %w", err)
	}

	h.log.Info("[HANDLER] Success! Updating status to 'sent'", "id", notification.ID)
	if err := h.statusUpdater.UpdateStatus(ctx, notification.ID, "sent"); err != nil {
		h.log.Error("failed to update status to sent", "id", notification.ID, "err", err)
		return fmt.Errorf("update status to sent: %w", err)
	}

	return nil
}
