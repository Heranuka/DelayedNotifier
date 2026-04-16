// Package cache provides cache helpers for application services.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Service provides caching operations.
type Service struct {
	cache Cache
}

// New creates a new cache service.
func New(cache Cache) *Service {
	return &Service{cache: cache}
}

// Set stores a value with expiration.
func (s *Service) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := s.cache.SetWithExpiration(ctx, key, value, expiration); err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}
	return nil
}

// GetString returns a cached string value.
func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	return s.cache.Get(ctx, key)
}

// GetJSON decodes a cached JSON value into dest.
func (s *Service) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a value from cache.
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.cache.Del(ctx, key)
}
