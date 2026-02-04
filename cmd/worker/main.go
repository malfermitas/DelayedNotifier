package main

import (
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/message_queue"
	"DelayedNotifier/internal/message_queue/message_queue_result"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/sender"
	"DelayedNotifier/internal/service"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	// 1. Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init Logger
	zlog.InitConsole()
	_ = zlog.SetLevel("info")

	// 3. Init Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)
	db, err := dbpg.New(dsn, nil, nil)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	emailSenderConfig := sender.EmailSenderConfig{
		SMTPHost: cfg.Email.SMTPHost,
		SMTPPort: cfg.Email.SMTPPort,
		SMTPUser: cfg.Email.Username,
		SMTPPass: cfg.Email.Password,
		From:     cfg.Email.From,
	}
	// 4. Init Senders
	emailSender, emailSenderInitErr := sender.NewEmailSender(emailSenderConfig)
	if emailSenderInitErr != nil {
		zlog.Logger.Error().Err(emailSenderInitErr).Msg("Failed to initialize email sender")
	}

	telegramSender, telegramSenderInitErr := sender.NewTelegramSender(cfg.Telegram.Token)
	go func() {
		telegramSender.Run()
		if telegramSenderInitErr != nil {
			zlog.Logger.Error().Err(telegramSenderInitErr).Msg("Failed to initialize telegram sender")
		}
	}()

	resultPublisher := message_queue_result.NewMessageQueueResultPublisher(
		cfg.RabbitMQ.URL,
		"notifier-worker-result-publisher",
		"notification_results_exchange",
		"notification.result",
	)

	// 5. Init Layers
	repo := repository.NewNotificationRepository(db, &zlog.Logger)
	svc := service.NewNotificationWorkerService(repo, resultPublisher, emailSender, telegramSender)

	// 6. Init and Start Consumer
	consumer := message_queue.NewMessageQueueConsumer(cfg.RabbitMQ.URL, "notifier-worker", svc)
	ctxConsumer, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	go func() {
		zlog.Logger.Info().Msg("Starting RabbitMQ consumer")
		err := <-consumer.Start(ctxConsumer)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("Consumer error")
		}
	}()

	go func() {
		zlog.Logger.Info().Msg("Starting RabbitMQ result publisher")
		err = resultPublisher.Start()
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("Consumer error")
		}
	}()

	// 7. Graceful Shutdown setup
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	zlog.Logger.Info().Msg("Graceful shutdown initiated...")

	// Cancel consumer context
	cancelConsumer()

	if err := consumer.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to close RabbitMQ consumer")
	}

	if err := resultPublisher.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to close result publisher")
	}

	zlog.Logger.Info().Msg("Worker exiting")
}
