package repository

import (
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type ChatIdsRepository interface {
	SetByUserID(ctx context.Context, userID string, chatID string) error
	GetByUserID(ctx context.Context, userID string) (string, error)
}

type chatIdsRepository struct {
	db *dbpg.DB
}

func NewChatIdsRepository(db *dbpg.DB) ChatIdsRepository {
	return chatIdsRepository{db: db}
}

func (c chatIdsRepository) SetByUserID(ctx context.Context, userID string, chatID string) error {
	query := `
		INSERT INTO telegram_chats (user_id, telegram_chat_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET telegram_chat_id = EXCLUDED.telegram_chat_id
	`

	_, err := c.db.ExecContext(ctx, query, userID, chatID)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("userID", userID).
			Str("chatID", chatID).
			Msg("Failed to set telegram chat by user id")
		return err
	}
	return nil
}

func (c chatIdsRepository) GetByUserID(ctx context.Context, userID string) (string, error) {
	query := `SELECT telegram_chat_id FROM telegram_chats WHERE user_id = $1`

	telegramChatId := ""

	err := c.db.QueryRowContext(ctx, query, userID).Scan(&telegramChatId)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("userID", userID).
			Msg("Failed to get telegram chat by user id")
		return "", err
	}

	return telegramChatId, nil
}
