// Package redis provides Redis client initialization helpers.
package redis

import (
	"delay/internal/config"

	"github.com/wb-go/wbf/redis"
)

// New creates a new Redis client using the provided configuration.
func New(cfg *config.Redis) *redis.Client {
	client := redis.New(cfg.Addr, cfg.Password, cfg.DB)
	return client
}
