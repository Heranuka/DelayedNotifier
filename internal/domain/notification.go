// Package domain defines the application's core business entities, value objects,
// and domain-level contracts.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a notification message to be delivered through a channel.
type Notification struct {
	ID          uuid.UUID           `json:"id"`
	Message     string              `json:"message"`
	Destination string              `json:"destination"`
	Channel     NotificationChannel `json:"channel"`
	Status      NotificationStatus  `json:"status"`
	DataToSent  time.Time           `json:"data_sent_at"`
	CreatedAt   time.Time           `json:"created_at"`
}
