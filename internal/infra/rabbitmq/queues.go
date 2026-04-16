package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupQueues declares the DLX, DLQ, binds the DLQ to the DLX,
// and declares the main queue with dead-letter settings.
func (c *Client) SetupQueues() error {
	if c == nil || c.inner == nil || c.cfg == nil {
		return fmt.Errorf("rabbitmq client is not initialized")
	}

	if c.cfg.DLXName == "" || c.cfg.DLQName == "" || c.cfg.QueueName == "" || c.cfg.ExchangeName == "" {
		return fmt.Errorf("rabbitmq queue names are required")
	}

	c.log.Info("[RABBIT] Declaring exchanges...", "exchange", c.cfg.ExchangeName, "dlx", c.cfg.DLXName)
	if err := c.inner.DeclareExchange(c.cfg.ExchangeName, "direct", true, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %q: %w", c.cfg.ExchangeName, err)
	}

	if err := c.inner.DeclareExchange(c.cfg.DLXName, "direct", true, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx %q: %w", c.cfg.DLXName, err)
	}

	c.log.Info("[RABBIT] Declaring queues...", "queue", c.cfg.QueueName, "dlq", c.cfg.DLQName)
	if err := c.inner.DeclareQueue(c.cfg.DLQName, c.cfg.DLXName, c.cfg.DLQName, true, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq %q: %w", c.cfg.DLQName, err)
	}

	mainQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    c.cfg.DLXName,
		"x-dead-letter-routing-key": c.cfg.DLQName,
	}

	if err := c.inner.DeclareQueue(c.cfg.QueueName, c.cfg.ExchangeName, c.cfg.QueueName, true, false, false, mainQueueArgs); err != nil {
		return fmt.Errorf("declare main queue %q: %w", c.cfg.QueueName, err)
	}

	// Explicitly bind the queue to the exchange just in case DeclareQueue doesn't do it.
	c.log.Info("[RABBIT] Binding queue to exchange...", "queue", c.cfg.QueueName, "exchange", c.cfg.ExchangeName)
	ch, err := c.GetChannel()
	if err != nil {
		return fmt.Errorf("get channel for binding: %w", err)
	}
	defer func() { _ = ch.Close() }()

	if err := ch.QueueBind(c.cfg.QueueName, c.cfg.QueueName, c.cfg.ExchangeName, false, nil); err != nil {
		return fmt.Errorf("bind queue %q to exchange %q: %w", c.cfg.QueueName, c.cfg.ExchangeName, err)
	}

	c.log.Info("[RABBIT] RabbitMQ setup completed successfully")
	return nil
}
