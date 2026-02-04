package service

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/shared"
	"context"
	"errors"
	"time"

	"github.com/wb-go/wbf/zlog"
)

type NotificationWorkerService struct {
	repo            repository.NotificationRepository
	resultPublisher shared.ResultPublisher
	emailSender     shared.Sender
	telegramSender  shared.Sender
}

func NewNotificationWorkerService(
	repo repository.NotificationRepository,
	resultPublisher shared.ResultPublisher,
	emailSender shared.Sender,
	telegramSender shared.Sender,
) *NotificationWorkerService {
	return &NotificationWorkerService{
		repo:            repo,
		resultPublisher: resultPublisher,
		emailSender:     emailSender,
		telegramSender:  telegramSender,
	}
}

func (s *NotificationWorkerService) ProcessNotificationFromQueue(ctx context.Context, notification model.Notification) error {
	zlog.Logger.Info().
		Str("notification_id", notification.ID).
		Str("channel", string(notification.Channel)).
		Msg("Processing notification from queue")

	notificationDB, err := s.repo.GetByID(ctx, notification.ID)
	if err != nil {
		zlog.Logger.Error().Str("notification_id", notification.ID).Msg("Failed to get notification from DB")
		return err
	}

	if notificationDB.Status == model.StatusCancelled {
		zlog.Logger.Info().Str("notification_id", notification.ID).Msg("Notification already cancelled")
		return nil
	}

	// Send notification based on channel
	startTime := time.Now()
	err = s.sendNotification(ctx, notification)
	duration := time.Since(startTime)

	// Create result
	result := model.NotificationResult{
		ID:     notification.ID,
		SentAt: time.Now(),
	}

	if err != nil {
		result.Status = model.StatusFailed
		result.Error = err.Error()
		zlog.Logger.Error().
			Err(err).
			Str("notification_id", notification.ID).
			Dur("duration", duration).
			Msg("Failed to send notification")
	} else {
		result.Status = model.StatusSent
		zlog.Logger.Info().
			Str("notification_id", notification.ID).
			Dur("duration", duration).
			Msg("Notification sent successfully")
	}

	// Publish result
	publishErr := s.resultPublisher.PublishResult(result)
	if publishErr != nil {
		zlog.Logger.Error().
			Err(publishErr).
			Str("notification_id", notification.ID).
			Msg("Failed to publish notification result")
	}

	// Always return nil to acknowledge the message
	return nil
}

func (s *NotificationWorkerService) sendNotification(ctx context.Context, notification model.Notification) error {
	switch notification.Channel {
	case "email":
		if s.emailSender == nil {
			return errors.New("email sender not configured")
		}
		return s.emailSender.Send(ctx, notification.Email, notification.Message)
	case "telegram":
		if s.telegramSender == nil {
			return errors.New("telegram sender not configured")
		}
		return s.telegramSender.Send(ctx, notification.TelegramID, notification.Message)
	default:
		zlog.Logger.Error().
			Str("channel", string(notification.Channel)).
			Str("notification_id", notification.ID).
			Msg("Unknown notification channel")
		return errors.New("unknown channel")
	}
}
