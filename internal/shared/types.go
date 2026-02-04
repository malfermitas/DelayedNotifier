package shared

import (
	"DelayedNotifier/internal/model"
	"context"
)

type NotificationQueueProcessor interface {
	ProcessNotificationFromQueue(ctx context.Context, notification model.Notification) error
}

type NotificationService interface {
	CreateNotification(ctx context.Context, message, sendAt, channel string) (string, error)
	GetNotificationById(ctx context.Context, id string) (*model.Notification, error)
	GetAllNotifications(ctx context.Context) ([]*model.Notification, error)
	DeleteNotificationById(ctx context.Context, id string) error
	MarkNotificationAsCancelled(ctx context.Context, id string) error
	ProcessNotificationResult(ctx context.Context, result model.NotificationResult) error
}

type MessageQueuePublisher interface {
	PublishNotification(ctx context.Context, notification *model.Notification) error
	Start() error
	Close() error
}

type MessageQueueConsumer interface {
	Start(ctx context.Context) <-chan error
	Close() error
}

type ResultPublisher interface {
	PublishResult(result model.NotificationResult) error
}
