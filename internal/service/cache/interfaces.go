// Package cache provides cache abstractions used by the application.
package cache

import (
	"context"
	"time"
)

// Cache defines the cache operations used by the application.
type Cache interface {
	SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}
