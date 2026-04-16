package notifications

import (
	"context"
	"delay/internal/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type Service interface {
	Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error)
	Status(ctx context.Context, noteID uuid.UUID) (string, error)
	Cancel(ctx context.Context, noteID uuid.UUID) error
	GetAll(ctx context.Context) (*[]domain.Notification, error)
}
