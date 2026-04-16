// Package notifications provides notification business logic.
package notifications

import (
	"context"
	"delay/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Cache defines the cache operations used by the notifications service.
type Cache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	GetJSON(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
}

// Publisher defines the message publishing operations used by the notifications service.
type Publisher interface {
	Publish(ctx context.Context, note *domain.Notification) error
}

// Repository defines the persistence operations used by the notifications service.
type Repository interface {
	Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error)
	Status(ctx context.Context, id uuid.UUID) (string, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) (*[]domain.Notification, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}
