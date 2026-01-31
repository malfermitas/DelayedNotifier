package shared

import (
	"DelayedNotifier/internal/model"
	"context"
)

type NotificationQueueProcessor interface {
	ProcessNotificationFromQueue(ctx context.Context, notification model.Notification) error
}
