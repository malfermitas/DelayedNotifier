package message_queue_result

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/shared"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type MessageQueueResultConsumer struct {
	consumer *rabbitmq.Consumer
	service  shared.NotificationService // Теперь используем интерфейс сервиса
	client   *rabbitmq.RabbitClient
}

func NewMessageQueueResultConsumer(url, connectionName string, service shared.NotificationService) *MessageQueueResultConsumer {
	config := rabbitmq.ClientConfig{
		URL:            url,
		ConnectionName: connectionName,
		ConnectTimeout: 0,
		Heartbeat:      0,
		ReconnectStrat: retry.Strategy{Attempts: 10, Delay: 5, Backoff: 1},
		ProducingStrat: retry.Strategy{Attempts: 3, Delay: 1, Backoff: 2},
		ConsumingStrat: retry.Strategy{Attempts: 3, Delay: 1, Backoff: 2},
	}

	rabbitClient, err := rabbitmq.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RabbitMQ client: %v", err))
	}

	consumerConfig := rabbitmq.ConsumerConfig{
		Queue:         "notification_results_queue",
		ConsumerTag:   "notification_result_consumer",
		AutoAck:       false,
		Ask:           rabbitmq.AskConfig{Multiple: false},
		Nack:          rabbitmq.NackConfig{Multiple: false, Requeue: true},
		Args:          nil,
		Workers:       1,
		PrefetchCount: 10,
	}

	queueConsumer := &MessageQueueResultConsumer{}

	consumer := rabbitmq.NewConsumer(rabbitClient, consumerConfig, func(ctx context.Context, delivery amqp091.Delivery) error {
		zlog.Logger.Debug().Str("body", string(delivery.Body)).Msg("Received notification result")

		result := model.NotificationResult{}
		jsonErr := json.Unmarshal(delivery.Body, &result)
		if jsonErr != nil {
			zlog.Logger.Err(jsonErr).Str("body", string(delivery.Body)).Msg("Failed to unmarshal notification result")
			return jsonErr
		}

		// Обрабатываем результат
		processResultError := service.ProcessNotificationResult(ctx, result)
		if processResultError != nil {
			zlog.Logger.Error().
				Err(processResultError).
				Str("notification_id", result.ID).
				Msg("Failed to process notification result")
			return processResultError
		}

		return nil
	})

	queueConsumer.consumer = consumer
	queueConsumer.service = service
	queueConsumer.client = rabbitClient

	return queueConsumer
}

func (c *MessageQueueResultConsumer) Start(ctx context.Context) <-chan error {
	chanError := make(chan error)
	go func() {
		err := c.client.DeclareExchange(
			"notification_results_exchange",
			"topic",
			true,
			false,
			false,
			nil,
		)
		if err != nil {
			chanError <- fmt.Errorf("failed to declare exchange: %w", err)
			return
		}

		err = c.client.DeclareQueue(
			"notification_results_queue",
			"notification_results_exchange",
			"notification.result",
			true,
			false,
			false,
			nil,
		)
		if err != nil {
			chanError <- fmt.Errorf("failed to declare queue: %w", err)
			return
		}

		chanError <- c.consumer.Start(ctx)
	}()
	return chanError
}

func (c *MessageQueueResultConsumer) Close() error {
	return c.client.Close()
}
