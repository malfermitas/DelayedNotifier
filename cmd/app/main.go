package main

import (
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/delivery"
	"DelayedNotifier/internal/delivery/handler"
	"DelayedNotifier/internal/message_queue"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/service"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 4. Init RabbitMQ Publisher
	publisher := message_queue.NewMessageQueuePublisher(
		cfg.RabbitMQ.URL,
		"notifier-publisher",
		"delayed_exchange",
		"notifications_key",
	)

	// 5. Init Layers
	repo := repository.NewNotificationRepository(db, &zlog.Logger)
	svc := service.NewNotificationService(repo, publisher)
	h := handler.NewNotificationHandler(svc)
	router := delivery.NewRouter(h)

	// 6. Init and Start Consumer
	consumer := message_queue.NewMessageQueueConsumer(cfg.RabbitMQ.URL, "notifier-consumer", svc)
	ctxConsumer, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	go func() {
		zlog.Logger.Info().Msg("Starting RabbitMQ producer")
		if err := publisher.Start(); err != nil {
			zlog.Logger.Error().Err(err).Msg("Publisher error")
		}
	}()

	go func() {
		zlog.Logger.Info().Msg("Starting RabbitMQ consumer")
		err := <-consumer.Start(ctxConsumer)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("Consumer error")
		}
	}()

	// 7. HTTP Server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// 8. Graceful Shutdown setup
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		zlog.Logger.Info().Msgf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	<-done
	zlog.Logger.Info().Msg("Graceful shutdown initiated...")

	// 9. Shutdown process
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	cancelConsumer()

	if err := server.Shutdown(ctxShutdown); err != nil {
		zlog.Logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	if err := publisher.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to close RabbitMQ publisher")
	}

	zlog.Logger.Info().Msg("Server exiting")
}
