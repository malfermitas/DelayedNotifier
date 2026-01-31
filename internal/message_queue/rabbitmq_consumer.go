package message_queue

import (
	"DelayedNotifier/internal/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/shared"
)

type MessageQueueConsumer struct {
	consumer *rabbitmq.Consumer
	service  *shared.NotificationQueueProcessor
}

func NewMessageQueueConsumer(url, connectionName string, processor shared.NotificationQueueProcessor) *MessageQueueConsumer {
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

	consumerConfig := rabbitmq.ConsumerConfig{
		Queue:         "notifications_queue",
		ConsumerTag:   "notification_consumer",
		AutoAck:       false,
		Ask:           rabbitmq.AskConfig{},
		Nack:          rabbitmq.NackConfig{},
		Args:          nil,
		Workers:       1,
		PrefetchCount: 10,
	}

	queueConsumer := &MessageQueueConsumer{}

	consumer := rabbitmq.NewConsumer(rabbitClient, consumerConfig, func(ctx context.Context, delivery amqp091.Delivery) error {
		zlog.Logger.Debug().Str("body", string(delivery.Body)).Msg("Received notification")

		notification := model.Notification{}
		jsonErr := json.Unmarshal(delivery.Body, &notification)
		if jsonErr != nil {
			zlog.Logger.Err(jsonErr).Str("body", string(delivery.Body)).Msg("Failed to unmarshal notification")
			return jsonErr
		}

		processNotificationError := processor.ProcessNotificationFromQueue(ctx, notification)
		if processNotificationError != nil {
			_ = delivery.Nack(false, true)
			return processNotificationError
		}

		err := delivery.Ack(false)
		if err != nil {
			fmt.Printf("Failed to ack message: %v\n", err)
			return err
		}
		return nil
	})

	queueConsumer.consumer = consumer
	queueConsumer.service = &processor

	return queueConsumer
}

func (c *MessageQueueConsumer) Start(ctx context.Context) error {
	return c.consumer.Start(ctx)
}
