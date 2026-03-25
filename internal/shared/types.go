package shared

import (
	"DelayedNotifier/internal/model"
	"context"
)

type Sender interface {
	Send(ctx context.Context, to string, message string) error
}

type NotificationQueueProcessor interface {
	ProcessNotificationFromQueue(ctx context.Context, notification model.Notification) error
}

type NotificationService interface {
	CreateNotification(ctx context.Context, message, sendAt, channel, email, telegramChatID, userID string) (string, error)
	GetNotificationById(ctx context.Context, id string) (*model.Notification, error)
	GetAllNotifications(ctx context.Context) ([]*model.Notification, error)
	DeleteNotificationById(ctx context.Context, id string) error
	MarkNotificationAsCancelled(ctx context.Context, id string) error
	ProcessNotificationResult(ctx context.Context, result model.NotificationResult) error
	ReadChatData(ctx context.Context, chatID int64, userID string)
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

type TelegramChatIDReader interface {
	ReadChatData(ctx context.Context, chatID int64, userID string)
}
