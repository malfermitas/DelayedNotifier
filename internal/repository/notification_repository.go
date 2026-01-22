package repository

import (
	"DelayedNotifier/internal/model"
	"context"
	"database/sql"
	"errors"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type NotificationRepository interface {
	Save(notification *model.Notification) error
	GetByID(id string) (*model.Notification, error)
	Remove(id string) error
}

type notificationRepository struct {
	db     *dbpg.DB
	logger *zlog.Zerolog
}

func NewNotificationRepository(db *dbpg.DB, logger *zlog.Zerolog) NotificationRepository {
	return &notificationRepository{
		db:     db,
		logger: logger,
	}
}

func (n notificationRepository) Save(notification *model.Notification) error {
	query := `
		INSERT INTO notifications (id, message, send_at, status, channel, created_at, updated_at, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			message = EXCLUDED.message,
			send_at = EXCLUDED.send_at,
			status = EXCLUDED.status,
			channel = EXCLUDED.channel,
			updated_at = EXCLUDED.updated_at,
			retry_count = EXCLUDED.retry_count
	`
	_, err := n.db.ExecContext(context.Background(), query,
		notification.ID,
		notification.Message,
		notification.SendAt,
		notification.Status,
		notification.Channel,
		notification.CreatedAt,
		notification.UpdatedAt,
		notification.RetryCount,
	)

	if err != nil {
		n.logger.Error().Str("id", notification.ID).Err(err).Msg("Error inserting notification")
		return err
	}

	n.logger.Info().Str("id", notification.ID).Msg("Notification saved")
	return nil
}

func (n notificationRepository) GetByID(id string) (*model.Notification, error) {
	query := `
		SELECT id, message, send_at, status, channel, created_at, updated_at, retry_count
		FROM notifications 
		WHERE id = $1
	`

	notification := &model.Notification{}

	err := n.db.QueryRowContext(context.Background(), query, id).Scan(
		&notification.ID,
		&notification.Message,
		&notification.SendAt,
		&notification.Status,
		&notification.Channel,
		&notification.CreatedAt,
		&notification.UpdatedAt,
		&notification.RetryCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			n.logger.Warn().Str("id", id).Msg("notification not found")
			return nil, err
		}
		n.logger.Error().Str("id", id).Str("error", err.Error()).Msg("failed to get notification by ID")
		return nil, err
	}

	n.logger.Info().Str("id", id).Msg("notification retrieved")

	return notification, nil
}

func (n notificationRepository) Remove(id string) error {
	query := "DELETE FROM notifications WHERE id = $1"
	result, err := n.db.ExecContext(context.Background(), query, id)
	if err != nil {
		n.logger.Error().Str("id", id).Str("error", err.Error()).Msg("failed to remove notification")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil && rowsAffected == 0 {
		n.logger.Error().Str("id", id).Str("error", err.Error()).Msg("no notification found to remove")
		return sql.ErrNoRows
	}

	n.logger.Info().Str("id", id).Msg("notification removed")
	return nil
}
