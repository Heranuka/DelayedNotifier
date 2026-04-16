package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

// ChannelProvider provides RabbitMQ channels.
type ChannelProvider interface {
	GetChannel() (*amqp.Channel, error)
}
