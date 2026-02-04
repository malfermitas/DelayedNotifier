package message_queue

import (
	"DelayedNotifier/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
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
		ReconnectStrat: retry.Strategy{Attempts: attemptsInf, Delay: 5, Backoff: 1},
		ProducingStrat: retry.Strategy{Attempts: 1, Delay: 10, Backoff: 2},
		ConsumingStrat: retry.Strategy{Attempts: 1, Delay: 10, Backoff: 2},
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

func (p *MessageQueuePublisher) Start() error {
	err := p.client.DeclareExchange(
		"delayed_exchange",
		"x-delayed-message",
		true,
		true,
		false,
		amqp091.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to declare exchange")
		return err
	}

	err = p.client.DeclareQueue(
		"notifications_queue",
		"delayed_exchange",
		p.routingKey,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to declare queue")
		return err
	}

	return nil
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
		opts = append(opts, rabbitmq.WithHeaders(amqp091.Table{"x-delay": int(delay.Milliseconds())}))
	}

	return p.publisher.Publish(
		ctx,
		body,
		p.routingKey,
		opts...,
	)
}

func (p *MessageQueuePublisher) PublishNotification(ctx context.Context, notification *model.Notification) error {
	delay := time.Until(notification.SendAt)
	if delay < 0 {
		delay = 0
	}

	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	err = p.PublishDelayed(ctx, body, delay)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to publish notification")
		return err
	}

	zlog.Logger.Info().Str("notification_id", notification.ID).Msg("Notification published")
	return nil
}

func (p *MessageQueuePublisher) Close() error {
	return p.client.Close()
}
