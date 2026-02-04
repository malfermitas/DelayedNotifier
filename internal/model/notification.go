package model

import (
	"time"
)

type NotificationChannel string

const (
	EmailChannel    NotificationChannel = "email"
	TelegramChannel NotificationChannel = "telegram"
)

func (ch NotificationChannel) IsValidChannelName() bool {
	switch ch {
	case EmailChannel, TelegramChannel:
		return true
	}
	return false
}

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusScheduled NotificationStatus = "scheduled"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusCancelled NotificationStatus = "cancelled"
)

type Notification struct {
	ID         string              `json:"id" db:"id"`
	Message    string              `json:"message" db:"message"`
	SendAt     time.Time           `json:"send_at" db:"send_at"`
	Status     NotificationStatus  `json:"status" db:"status"`
	Channel    NotificationChannel `json:"channel" db:"channel"`
	Email      string              `json:"email" db:"email"`
	TelegramID string              `json:"telegram_id" db:"telegram_id"`
	CreatedAt  time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at" db:"updated_at"`
	RetryCount int                 `json:"retry_count" db:"retry_count"`
}

func NewNotification(
	id,
	message string,
	sendAt time.Time,
	channel NotificationChannel,
	email string,
	telegramChatID string,
) *Notification {
	now := time.Now()
	return &Notification{
		ID:         id,
		Message:    message,
		SendAt:     sendAt,
		Status:     StatusPending,
		Channel:    channel,
		Email:      email,
		TelegramID: telegramChatID,
		CreatedAt:  now,
		UpdatedAt:  now,
		RetryCount: 0,
	}
}
