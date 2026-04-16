package notifications

import (
	"delay/internal/domain"
	"time"
)

type createRequest struct {
	Message     string                     `json:"message"`
	Destination string                     `json:"destination"`
	Channel     domain.NotificationChannel `json:"channel"`
	DataToSent  time.Time                  `json:"data_sent_at"`
}

type notificationResponse struct {
	ID          string `json:"id"`
	Destination string `json:"destination"`
	Channel     string `json:"channel"`
	Message     string `json:"message"`
	Status      string `json:"status"`
	DataSentAt  string `json:"data_sent_at"`
	CreatedAt   string `json:"created_at"`
}
