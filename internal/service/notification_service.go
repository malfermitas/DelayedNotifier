package service

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/wb-go/wbf/helpers"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, message, sendAt, channel string) (string, error)
	GetNotificationById(ctx context.Context, id string) (*model.Notification, error)
	DeleteNotificationById(ctx context.Context, id string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return notificationService{repo: repo}
}

// CreateNotification creates a new notification with the provided message, send time, and channel.
// It returns the ID of the created notification or an error if the operation fails.
func (n notificationService) CreateNotification(
	ctx context.Context,
	message,
	sendAt,
	channel string,
) (string, error) {
	notificationId := helpers.CreateUUID()
	sendAtTime := time.Now()
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

	err := n.repo.Save(notification)
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
