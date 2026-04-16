// Package rabbitmq provides RabbitMQ connection and queue setup helpers.
package rabbitmq

import (
	"context"
	"delay/internal/config"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/logger"
	wbrabbitmq "github.com/wb-go/wbf/rabbitmq"
)

type Client struct {
	inner  *wbrabbitmq.RabbitClient
	cancel context.CancelFunc
	wg     sync.WaitGroup
	cfg    *config.RabbitMQ
	log    *logger.ZerologAdapter
}

// NewClient connects to RabbitMQ and declares required queues.
func NewClient(cfg *config.RabbitMQ, log *logger.ZerologAdapter) (*Client, error) {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	inner, err := wbrabbitmq.NewClient(wbrabbitmq.ClientConfig{
		URL: url,
	})
	if err != nil {
		return nil, err
	}

	client := &Client{
		inner: inner,
		cfg:   cfg,
		log:   log,
	}

	return client, nil
}

func (c *Client) GetChannel() (*amqp.Channel, error) {
	if c == nil || c.inner == nil {
		return nil, fmt.Errorf("rabbitmq client is not initialized")
	}
	return c.inner.GetChannel()
}

func (c *Client) Stop() error {
	if c == nil {
		return nil
	}

	c.log.Info("Stopping RabbitMQ...")
	defer c.log.Info("RabbitMQ stopped")

	if c.cancel != nil {
		c.cancel()
	}

	c.wg.Wait()

	if c.inner != nil {
		if err := c.inner.Close(); err != nil {
			c.log.Error("Error closing rabbitmq client")
			return err
		}
	}

	return nil
}
