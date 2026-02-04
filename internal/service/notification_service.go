package service

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/shared"
	"context"
	"errors"
	"time"

	"github.com/wb-go/wbf/helpers"
	"github.com/wb-go/wbf/zlog"
)

type notificationService struct {
	repo            repository.NotificationRepository
	publisher       shared.MessageQueuePublisher
	resultPublisher shared.ResultPublisher
}

func NewNotificationService(
	repo repository.NotificationRepository,
	publisher shared.MessageQueuePublisher,
) shared.NotificationService {
	return notificationService{
		repo:      repo,
		publisher: publisher,
	}
}

var NotificationCancelledError = errors.New("notification cancelled")

// CreateNotification creates a new notification with the provided message, send time, and channel.
// It returns the ID of the created notification or an error if the operation fails.
func (n notificationService) CreateNotification(
	ctx context.Context,
	message,
	sendAt,
	channel string,
) (string, error) {
	notificationId := helpers.CreateUUID()
	sendAtTime, conversionError := time.Parse(time.RFC3339, sendAt)
	if conversionError != nil {
		// Fallback to old format if needed, but RFC3339 is preferred now
		sendAtTime, conversionError = time.Parse("2006-01-02 15:04:05", sendAt)
		if conversionError != nil {
			return "", conversionError
		}
	}

	notificationChannel := model.NotificationChannel(channel)

	if !notificationChannel.IsValidChannelName() {
		return "", errors.New("invalid notification channel name")
	}

	notification := model.NewNotification(
		notificationId,
		message,
		sendAtTime,
		notificationChannel,
	)

	notification.Status = model.StatusPending

	err := n.repo.Save(notification)
	if err != nil {
		return "", err
	}

	err = n.publisher.PublishNotification(ctx, notification)
	if err != nil {
		_ = n.repo.UpdateStatus(notificationId, model.StatusFailed)
		return "", err
	}

	err = n.repo.UpdateStatus(notification.ID, model.StatusScheduled)
	if err != nil {
		return "", err
	}

	return notification.ID, nil
}

func (n notificationService) GetNotificationById(
	ctx context.Context,
	id string,
) (*model.Notification, error) {
	notification, err := n.repo.GetByID(id)
	return notification, err
}

func (n notificationService) GetAllNotifications(ctx context.Context) ([]*model.Notification, error) {
	return n.repo.GetAll()
}

func (n notificationService) DeleteNotificationById(ctx context.Context, id string) error {
	err := n.repo.Remove(id)
	return err
}

func (n notificationService) MarkNotificationAsCancelled(ctx context.Context, id string) error {
	return n.repo.UpdateStatus(id, model.StatusCancelled)
}

func (n notificationService) ProcessNotificationResult(ctx context.Context, result model.NotificationResult) error {
	err := n.repo.UpdateStatus(result.ID, result.Status)
	if err != nil {
		zlog.Logger.Error().Str("notification_id", result.ID).Str("status", string(result.Status))
		return err
	}
	zlog.Logger.Info().Str("notification_id", result.ID).Msg("notification processed")
	return nil
}
