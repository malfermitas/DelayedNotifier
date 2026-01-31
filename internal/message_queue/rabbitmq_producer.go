package message_queue

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

type MessageQueuePublisher struct {
	client     *rabbitmq.RabbitClient
	publisher  *rabbitmq.Publisher
	routingKey string
}

func NewMessageQueuePublisher(url, connectionName, exchange, routingKey string) *MessageQueuePublisher {
	cfg := rabbitmq.ClientConfig{
		URL:            url,
		ConnectionName: connectionName,
		ConnectTimeout: 0,
		Heartbeat:      0,
		ReconnectStrat: retry.Strategy{},
		ProducingStrat: retry.Strategy{},
		ConsumingStrat: retry.Strategy{},
	}

	rabbitClient, err := rabbitmq.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RabbitMQ client: %v", err))
	}

	publisher := rabbitmq.NewPublisher(rabbitClient, exchange, routingKey)

	return &MessageQueuePublisher{
		client:     rabbitClient,
		publisher:  publisher,
		routingKey: routingKey,
	}
}

func (p *MessageQueuePublisher) Publish(ctx context.Context, body []byte) error {
	return p.publisher.Publish(
		ctx,
		body,
		p.routingKey,
	)
}

func (p *MessageQueuePublisher) PublishDelayed(ctx context.Context, body []byte, delay time.Duration) error {
	opts := make([]rabbitmq.PublishOption, 0, 1)
	if delay > 0 {
		opts = append(opts, rabbitmq.WithHeaders(amqp091.Table{"x-delay": delay.Milliseconds()}))
	}

	return p.publisher.Publish(
		ctx,
		body,
		p.routingKey,
		opts...,
	)
}

func (p *MessageQueuePublisher) PublishNotificationID(ctx context.Context, notificationID string, sendAt time.Time) error {
	delay := time.Until(sendAt)
	if delay < 0 {
		delay = 0
	}

	return p.PublishDelayed(ctx, []byte(notificationID), delay)
}

func (p *MessageQueuePublisher) Close() error {
	return p.client.Close()
}
