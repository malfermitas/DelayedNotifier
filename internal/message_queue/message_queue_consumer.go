package message_queue

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/service"
	"DelayedNotifier/internal/shared"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type messageQueueConsumer struct {
	consumer *rabbitmq.Consumer
	// Store processor as interface value, not a pointer to interface
	service shared.NotificationQueueProcessor
	client  *rabbitmq.RabbitClient
}

const attemptsInf int = 1 << 32

func NewMessageQueueConsumer(url, connectionName string, processor shared.NotificationQueueProcessor) shared.MessageQueueConsumer {
	config := rabbitmq.ClientConfig{
		URL:            url,
		ConnectionName: connectionName,
		ConnectTimeout: 0,
		Heartbeat:      0,
		ReconnectStrat: retry.Strategy{Attempts: attemptsInf, Delay: 5, Backoff: 1},
		ProducingStrat: retry.Strategy{Attempts: 1, Delay: 10, Backoff: 2},
		ConsumingStrat: retry.Strategy{Attempts: 1, Delay: 10, Backoff: 2},
	}

	rabbitClient, err := rabbitmq.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RabbitMQ client: %v", err))
	}

	consumerConfig := rabbitmq.ConsumerConfig{
		Queue:         "notifications_queue",
		ConsumerTag:   "notification_consumer",
		AutoAck:       false,
		Ask:           rabbitmq.AskConfig{Multiple: false},
		Nack:          rabbitmq.NackConfig{Multiple: false, Requeue: true},
		Args:          nil,
		Workers:       1,
		PrefetchCount: 10,
	}

	queueConsumer := &messageQueueConsumer{}

	consumer := rabbitmq.NewConsumer(rabbitClient, consumerConfig, func(ctx context.Context, delivery amqp091.Delivery) error {
		zlog.Logger.Debug().Str("body", string(delivery.Body)).Msg("Received notification")

		notification := model.Notification{}
		jsonErr := json.Unmarshal(delivery.Body, &notification)
		if jsonErr != nil {
			zlog.Logger.Err(jsonErr).Str("body", string(delivery.Body)).Msg("Failed to unmarshal notification")
			return jsonErr
		}

		processNotificationError := processor.ProcessNotificationFromQueue(ctx, notification)
		if processNotificationError != nil &&
			!errors.Is(processNotificationError, service.NotificationCancelledError) {
			return processNotificationError
		}

		return nil
	})

	queueConsumer.consumer = consumer
	queueConsumer.service = processor
	queueConsumer.client = rabbitClient

	return queueConsumer
}

func (c *messageQueueConsumer) Start(ctx context.Context) <-chan error {
	chanError := make(chan error)
	go func() {
		err := c.client.DeclareExchange(
			"delayed_exchange",
			"x-delayed-message",
			true,
			true,
			false,
			amqp091.Table{"x-delayed-type": "direct"},
		)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to declare exchange")
			chanError <- err
			return
		}

		err = c.client.DeclareQueue(
			"notifications_queue",
			"delayed_exchange",
			"notifications_key",
			true,
			false,
			false,
			nil,
		)
		if err != nil {
			chanError <- err
			return
		}
		chanError <- c.consumer.Start(ctx)
	}()
	return chanError
}

func (c *messageQueueConsumer) Close() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	return nil
}
