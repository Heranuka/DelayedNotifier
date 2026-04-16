// Package notifications provides notification business logic.
package notifications

import (
	"context"
	"delay/internal/config"
	"delay/internal/domain"
	notificationsrepo "delay/internal/repository/notifications"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/logger"
)

// Service implements notification-related business operations.
type Service struct {
	repo      Repository
	publisher Publisher
	cache     Cache
	cfg       *config.Config
	logger    *logger.ZerologAdapter
}

// New creates a new Service instance.
func New(
	repo Repository,
	publisher Publisher,
	cache Cache,
	cfg *config.Config,
	logger *logger.ZerologAdapter,
) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		cache:     cache,
		cfg:       cfg,
		logger:    logger,
	}
}

// Create validates, persists, caches, and publishes a notification.
func (s *Service) Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error) {
	if notification == nil {
		s.logger.Warn("notification payload is nil")
		return uuid.Nil, ErrBadRequest
	}

	if notification.Message == "" {
		s.logger.Warn("notification message is required")
		return uuid.Nil, ErrMessageRequired
	}

	if !notification.DataToSent.IsZero() && notification.DataToSent.Before(time.Now().UTC()) {
		s.logger.Warn("notification delivery time must be in the future")
		return uuid.Nil, ErrDataSent
	}

	id, err := s.repo.Create(ctx, notification)
	if err != nil {
		s.logger.Error("failed to create notification in repository", "err", err)
		return uuid.Nil, err
	}

	if err := s.repo.UpdateStatus(ctx, id, "created"); err != nil {
		s.logger.Error("failed to update notification status", "note_id", id, "err", err)
		return uuid.Nil, err
	}

	note, err := s.Get(ctx, id)
	if err != nil {
		s.logger.Error("failed to load created notification", "note_id", id, "err", err)
		return uuid.Nil, err
	}

	cacheKey := "notification:status:" + id.String()
	if err := s.cache.Set(ctx, cacheKey, "created", 5*time.Minute); err != nil {
		s.logger.Error("failed to cache notification status", "note_id", id, "err", err)
	}

	if err := s.publisher.Publish(ctx, note); err != nil {
		s.logger.Error("failed to publish notification to RabbitMQ", "note_id", id, "err", err)
		return uuid.Nil, err
	}

	s.logger.Info("notification created successfully", "note_id", id, "channel", notification.Channel, "destination", notification.Destination)
	s.logger.Info("notification published successfully", "note_id", id)

	// Invalidate the "get all" cache
	_ = s.cache.Delete(ctx, "notification:getall")

	return id, nil
}

// Status returns notification status, using cache first and repository as fallback.
func (s *Service) Status(ctx context.Context, id uuid.UUID) (string, error) {
	cacheKey := "notification:status:" + id.String()
	var status string

	if status, err := s.cache.GetString(ctx, cacheKey); err == nil {
		s.logger.Debug("notification status cache hit", "note_id", id, "status", status)
		return status, nil
	}

	status, err := s.repo.Status(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.logger.Warn("notification status not found", "id", id)
			return "", ErrNotFound
		}

		s.logger.Error("failed to fetch notification status", "id", id, "err", err)
		return "", err
	}

	if err := s.cache.Set(ctx, cacheKey, status, 5*time.Minute); err != nil {
		s.logger.Error("failed to cache notification status", "id", id, "err", err)
	}

	s.logger.Info("notification status loaded successfully", "id", id, "status", status)
	return status, nil
}

// Cancel deletes a notification and clears its cached status.
func (s *Service) Cancel(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Cancel(ctx, id); err != nil {
		if errors.Is(err, notificationsrepo.ErrNotFound) {
			s.logger.Warn("notification not found for cancellation", "id", id)
			return ErrNotFound
		}

		s.logger.Error("failed to cancel notification", "id", id, "err", err)
		return err
	}

	// Invalidate the "get all" cache
	_ = s.cache.Delete(ctx, "notification:getall")

	cacheKey := "notification:status:" + id.String()
	if err := s.cache.Set(ctx, cacheKey, "", 0); err != nil {
		s.logger.Error("failed to clear notification status cache", "id", id, "err", err)
	}

	s.logger.Info("notification canceled successfully", "id", id)
	return nil
}

// GetAll returns all notifications, using cache first and repository as fallback.
func (s *Service) GetAll(ctx context.Context) (*[]domain.Notification, error) {
	cacheKey := "notification:getall"
	var notifications []domain.Notification

	if err := s.cache.GetJSON(ctx, cacheKey, &notifications); err == nil {
		s.logger.Debug("notifications cache hit", "count", len(notifications))
		return &notifications, nil
	}

	notificationsPtr, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to fetch notifications from repository", "err", err)
		return nil, err
	}

	notifications = *notificationsPtr

	bytes, err := json.Marshal(notifications)
	if err != nil {
		s.logger.Error("failed to marshal notifications for cache", "err", err)
		return &notifications, nil
	}

	if err := s.cache.Set(ctx, cacheKey, string(bytes), 5*time.Minute); err != nil {
		s.logger.Error("failed to cache notifications list", "err", err)
	}

	s.logger.Info("notifications loaded successfully", "count", len(notifications))
	return &notifications, nil
}

// Get returns a notification by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	note, err := s.repo.Get(ctx, id)
	if err != nil {
		if note == nil {
			s.logger.Warn("notification not found", "id", id)
			return nil, ErrNotFound
		}

		s.logger.Error("failed to get notification", "id", id, "err", err)
		return nil, err
	}

	s.logger.Debug("notification loaded successfully", "id", id)
	return note, nil
}

// UpdateStatus updates notification status.
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("failed to update notification status", "id", id, "status", status, "err", err)
		return err
	}

	// Invalidate the "get all" cache
	_ = s.cache.Delete(ctx, "notification:getall")

	cacheKey := "notification:status:" + id.String()
	if err := s.cache.Set(ctx, cacheKey, status, 5*time.Minute); err != nil {
		s.logger.Error("failed to update notification status cache", "id", id, "err", err)
	}

	s.logger.Info("notification status updated successfully", "id", id, "status", status)
	return nil
}
