package service

import (
	"DelayedNotifier/internal/model"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/shared"
	"context"
	"errors"
	"strconv"
	"strings"
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

var (
	ErrInvalidNotificationChannel = errors.New("invalid notification channel name")
	ErrEmailRecipientRequired     = errors.New("recipient email is required for email channel")
	ErrTelegramRecipientRequired  = errors.New("recipient chat_id or user_id is required for telegram channel")
	ErrTelegramRecipientNotFound  = errors.New("telegram chat_id not found for user_id")
)

// CreateNotification creates a new notification with the provided message, send time, and channel.
// It returns the ID of the created notification or an error if the operation fails.
// UserID is used to link telegram chat id to an external user.
func (n notificationService) CreateNotification(
	ctx context.Context,
	message,
	sendAt,
	channel string,
	email string,
	telegramChatID string,
	userID string,
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
		return "", ErrInvalidNotificationChannel
	}

	if notificationChannel == model.EmailChannel && email == "" {
		return "", ErrEmailRecipientRequired
	}

	if notificationChannel == model.TelegramChannel {
		if telegramChatID == "" {
			if userID == "" {
				return "", ErrTelegramRecipientRequired
			}

			resolvedChatID, err := n.chatIdsRepo.GetByUserID(ctx, userID)
			if err != nil {
				zlog.Logger.Error().
					Str("user_id", userID).
					Err(err).
					Msg("Failed to resolve telegram chat id by user id")
				return "", ErrTelegramRecipientNotFound
			}

			telegramChatID = resolvedChatID
		}
		email = ""
	} else {
		telegramChatID = ""
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

	err := n.notificationRepo.Save(ctx, notification)
	if err != nil {
		return "", err
	}

	err = n.publisher.PublishNotification(ctx, notification)
	if err != nil {
		_ = n.notificationRepo.UpdateStatus(ctx, notificationId, model.StatusFailed)
		return "", err
	}

	err = n.notificationRepo.UpdateStatus(ctx, notification.ID, model.StatusScheduled)
	if err != nil {
		return "", err
	}

	return notification.ID, nil
}

func (n notificationService) GetNotificationById(
	ctx context.Context,
	id string,
) (*model.Notification, error) {
	notification, err := n.notificationRepo.GetByID(ctx, id)
	return notification, err
}

func (n notificationService) GetAllNotifications(ctx context.Context) ([]*model.Notification, error) {
	return n.notificationRepo.GetAll(ctx)
}

func (n notificationService) DeleteNotificationById(ctx context.Context, id string) error {
	err := n.notificationRepo.Remove(ctx, id)
	return err
}

func (n notificationService) MarkNotificationAsCancelled(ctx context.Context, id string) error {
	err := n.notificationRepo.UpdateStatus(ctx, id, model.StatusCancelled)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("id", id).Msg("Failed to mark notification as cancelled")
		return err
	}
	zlog.Logger.Info().Str("id", id).Msg("Marked notification as cancelled")
	return nil
}

func (n notificationService) ProcessNotificationResult(ctx context.Context, result model.NotificationResult) error {
	err := n.notificationRepo.UpdateStatus(ctx, result.ID, result.Status)
	if err != nil {
		zlog.Logger.Error().Str("notification_id", result.ID).Str("status", string(result.Status))
		return err
	}
	zlog.Logger.Info().Str("notification_id", result.ID).Msg("notification processed")
	return nil
}

func (n notificationService) ReadChatData(ctx context.Context, chatID int64, userID string) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		zlog.Logger.Warn().Int64("chat_id", chatID).Msg("skip telegram binding because user_id is empty")
		return
	}

	zlog.Logger.Info().Int64("chat_id", chatID).Str("user_id", userID).Msg("read chat")
	err := n.chatIdsRepo.SetByUserID(ctx, userID, strconv.FormatInt(chatID, 10))
	if err != nil {
		zlog.Logger.Error().
			Int64("chat_id", chatID).
			Str("user_id", userID).
			Err(err).
			Msg("could not save chat id to database")
	}
}
