package service

import (
	"DelayedNotifier/internal/message_queue"
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/wb-go/wbf/helpers"
	"github.com/wb-go/wbf/zlog"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, message, sendAt, channel string) (string, error)
	GetNotificationById(ctx context.Context, id string) (*model.Notification, error)
	DeleteNotificationById(ctx context.Context, id string) error
	MarkNotificationAsCancelled(ctx context.Context, id string) error
	ProcessNotificationFromQueue(ctx context.Context, notification model.Notification) error
}
type notificationService struct {
	repo      repository.NotificationRepository
	publisher *message_queue.MessageQueuePublisher
}

func NewNotificationService(repo repository.NotificationRepository, publisher *message_queue.MessageQueuePublisher) NotificationService {
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
	sendAtTime, conversionError := time.Parse("2006-01-02 15:04:05", sendAt)
	if conversionError != nil {
		return "", conversionError
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

	err = n.publisher.PublishNotificationID(ctx, notification.ID, notification.SendAt)
	if err != nil {
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

func (n notificationService) DeleteNotificationById(ctx context.Context, id string) error {
	err := n.repo.Remove(id)
	return err
}

func (n notificationService) MarkNotificationAsCancelled(ctx context.Context, id string) error {
	return n.repo.UpdateStatus(id, model.StatusCancelled)
}

func (n notificationService) ProcessNotificationFromQueue(ctx context.Context, notification model.Notification) error {
	notificationFromDB, err := n.repo.GetByID(notification.ID)
	if err != nil {
		return err
	}

	if notificationFromDB.Status == model.StatusCancelled {
		return NotificationCancelledError
	}

	switch notification.Channel {
	case model.EmailChannel:
		// TODO: send email
	case model.TelegramChannel:
		// TODO: send telegram message
	}

	err = n.repo.UpdateStatus(notification.ID, model.StatusSent)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Str("id", notification.ID).Msg("Notification sent")

	return nil
}
