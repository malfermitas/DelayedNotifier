package message_queue_result

import (
	"DelayedNotifier/internal/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type MessageQueueResultPublisher struct {
	client     *rabbitmq.RabbitClient
	publisher  *rabbitmq.Publisher
	exchange   string
	routingKey string
}

func NewMessageQueueResultPublisher(url, connectionName, exchange, routingKey string) *MessageQueueResultPublisher {
	config := rabbitmq.ClientConfig{
		URL:            url,
		ConnectionName: connectionName,
		ConnectTimeout: 0,
		Heartbeat:      0,
		ReconnectStrat: retry.Strategy{Attempts: 10, Delay: 5, Backoff: 1},
		ProducingStrat: retry.Strategy{Attempts: 3, Delay: 1, Backoff: 2},
	}

	rabbitClient, err := rabbitmq.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RabbitMQ client: %v", err))
	}

	publisher := rabbitmq.NewPublisher(rabbitClient, exchange, routingKey)

	return &MessageQueueResultPublisher{
		client:     rabbitClient,
		publisher:  publisher,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (p *MessageQueueResultPublisher) Start() error {
	err := p.client.DeclareExchange(
		p.exchange,
		"topic",
		true,
		false,
		false,
		nil,
	)

	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to declare exchange")
		return err
	}

	err = p.client.DeclareQueue(
		"notification_results_queue",
		p.exchange,
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

func (p *MessageQueueResultPublisher) PublishResult(result model.NotificationResult) error {
	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal notification result: %w", err)
	}

	ctx := context.Background()
	err = p.publisher.Publish(ctx, body, "notification.result")
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("notification_id", result.ID).
			Msg("Failed to publish notification result")
		return err
	}

	zlog.Logger.Debug().
		Str("notification_id", result.ID).
		Str("status", string(result.Status)).
		Msg("Published notification result")

	return nil
}

func (p *MessageQueueResultPublisher) Close() error {
	return p.client.Close()
}
