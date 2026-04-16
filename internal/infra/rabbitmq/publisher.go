// Package rabbitmq provides RabbitMQ publisher and consumer adapters
// for notification delivery infrastructure.
package rabbitmq

import (
	"context"
	"delay/internal/domain"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/logger"
)

// Publisher publishes notifications to RabbitMQ.
type Publisher struct {
	client       ChannelProvider
	queueName    string
	exchangeName string
	log          *logger.ZerologAdapter
}

// NewPublisher creates a new RabbitMQ publisher.
func NewPublisher(client ChannelProvider, queueName, exchangeName string, log *logger.ZerologAdapter) *Publisher {
	return &Publisher{
		client:       client,
		queueName:    queueName,
		exchangeName: exchangeName,
		log:          log,
	}
}

// Publish marshals the notification and sends it to the RabbitMQ queue.
func (p *Publisher) Publish(ctx context.Context, note *domain.Notification) error {
	body, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ch, err := p.client.GetChannel()
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}
	defer func() { _ = ch.Close() }()

	p.log.Info("[RABBIT] Publishing message", "exchange", p.exchangeName, "routing_key", p.queueName)
	return ch.PublishWithContext(ctx, p.exchangeName, p.queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}
