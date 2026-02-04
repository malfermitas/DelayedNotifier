package service

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/shared"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/wb-go/wbf/helpers"
	"github.com/wb-go/wbf/zlog"
)

type notificationService struct {
	notificationRepo repository.NotificationRepository
	chatIdsRepo      repository.ChatIdsRepository
	publisher        shared.MessageQueuePublisher
}

func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	chatIdsRepo repository.ChatIdsRepository,
	publisher shared.MessageQueuePublisher,
) shared.NotificationService {
	return notificationService{
		notificationRepo: notificationRepo,
		chatIdsRepo:      chatIdsRepo,
		publisher:        publisher,
	}
}

var NotificationCancelledError = errors.New("notification cancelled")

// CreateNotification creates a new notification with the provided message, send time, and channel.
// It returns the ID of the created notification or an error if the operation fails.
// User cookie is used to link it to telegram chat id
func (n notificationService) CreateNotification(
	ctx context.Context,
	message,
	sendAt,
	channel string,
	email string,
	userCookie string,
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

	telegramChatID, err := n.chatIdsRepo.Get(userCookie)
	if err != nil {
		zlog.Logger.Error().
			Str("telegramChatID", telegramChatID).
			Str("cookie", userCookie).
			Err(err).
			Msg("Failed to get telegram chat id for cookie")

		return "", err
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
		email,
		telegramChatID,
	)

	notification.Status = model.StatusPending

	err = n.notificationRepo.Save(notification)
	if err != nil {
		return "", err
	}

	err = n.publisher.PublishNotification(ctx, notification)
	if err != nil {
		_ = n.notificationRepo.UpdateStatus(notificationId, model.StatusFailed)
		return "", err
	}

	err = n.notificationRepo.UpdateStatus(notification.ID, model.StatusScheduled)
	if err != nil {
		return "", err
	}

	return notification.ID, nil
}

func (n notificationService) GetNotificationById(
	ctx context.Context,
	id string,
) (*model.Notification, error) {
	notification, err := n.notificationRepo.GetByID(id)
	return notification, err
}

func (n notificationService) GetAllNotifications(ctx context.Context) ([]*model.Notification, error) {
	return n.notificationRepo.GetAll()
}

func (n notificationService) DeleteNotificationById(ctx context.Context, id string) error {
	err := n.notificationRepo.Remove(id)
	return err
}

func (n notificationService) MarkNotificationAsCancelled(ctx context.Context, id string) error {
	return n.notificationRepo.UpdateStatus(id, model.StatusCancelled)
}

func (n notificationService) ProcessNotificationResult(ctx context.Context, result model.NotificationResult) error {
	err := n.notificationRepo.UpdateStatus(result.ID, result.Status)
	if err != nil {
		zlog.Logger.Error().Str("notification_id", result.ID).Str("status", string(result.Status))
		return err
	}
	zlog.Logger.Info().Str("notification_id", result.ID).Msg("notification processed")
	return nil
}

func (n notificationService) ReadChatData(chatID int64, cookie string) {
	zlog.Logger.Info().Int64("chat_id", chatID).Str("cookie", cookie).Msg("read chat")
	err := n.chatIdsRepo.Set(cookie, strconv.FormatInt(chatID, 10))
	if err != nil {
		zlog.Logger.Error().
			Int64("chat_id", chatID).
			Str("cookie", cookie).
			Err(err).
			Msg("could not save chat id to database")
	}
}
