// Package rabbitmq provides RabbitMQ client, publisher, and consumer infrastructure.
package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/logger"
)

type Handler interface {
	Handle(ctx context.Context, body []byte) error
}

// Consumer consumes messages from a RabbitMQ queue and dispatches them to a handler.
type Consumer struct {
	ch          *amqp.Channel
	queue       string
	consumerTag string
	prefetch    int
	autoAck     bool
	handler     Handler
	log         *logger.ZerologAdapter
}

func NewConsumer(ch *amqp.Channel, queue string, handler Handler, log *logger.ZerologAdapter) *Consumer {
	return &Consumer{
		ch:          ch,
		queue:       queue,
		consumerTag: "simple-consumer",
		prefetch:    1,
		autoAck:     false,
		handler:     handler,
		log:         log,
	}
}

func (c *Consumer) WithConsumerTag(tag string) *Consumer {
	if tag != "" {
		c.consumerTag = tag
	}
	return c
}

func (c *Consumer) WithPrefetch(n int) *Consumer {
	if n > 0 {
		c.prefetch = n
	}
	return c
}

func (c *Consumer) Start(ctx context.Context) error {
	if c.ch == nil {
		return errors.New("rabbitmq channel is nil")
	}
	if c.handler == nil {
		return errors.New("message handler is nil")
	}
	if c.queue == "" {
		return errors.New("queue name is empty")
	}

	if err := c.ch.Qos(c.prefetch, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	c.log.Info("[RABBIT] Starting message consumption...", "queue", c.queue, "tag", c.consumerTag)
	deliveries, err := c.ch.Consume(
		c.queue,
		c.consumerTag,
		c.autoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.log.Info("[RABBIT] Consumer started and listening for messages")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("[RABBIT] Consumer stopping (context done)")
			return ctx.Err()

		case d, ok := <-deliveries:
			if !ok {
				c.log.Error("[RABBIT] Deliveries channel closed. Consumer stopping.")
				return nil
			}

			if err := c.handleDelivery(ctx, d); err != nil {
				c.log.Error("consumer delivery error", "err", err)
			}
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery) error {
	c.log.Info("[RABBIT] Received message", "body", string(d.Body))
	if err := c.handler.Handle(ctx, d.Body); err != nil {
		c.log.Error("[RABBIT] Handler error", "err", err)
		isFatal := true
		if nackErr := d.Nack(false, !isFatal); nackErr != nil {
			c.log.Error("[RABBIT] Nack error", "err", nackErr)
			return fmt.Errorf("handler error: %w; nack error: %v", err, nackErr)
		}
		return fmt.Errorf("handler error: %w", err)
	}

	if err := d.Ack(false); err != nil {
		c.log.Error("[RABBIT] Ack error", "err", err)
		return fmt.Errorf("ack error: %w", err)
	}

	c.log.Info("[RABBIT] Message processed and Acked successfully")
	return nil
}
