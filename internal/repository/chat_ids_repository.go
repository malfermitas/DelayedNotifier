package repository

import (
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type ChatIdsRepository interface {
	Set(cookieID string, chatID string) error
	Get(cookieID string) (string, error)
}

type chatIdsRepository struct {
	db *dbpg.DB
}

func NewChatIdsRepository(db *dbpg.DB) ChatIdsRepository {
	return chatIdsRepository{db: db}
}

func (c chatIdsRepository) Set(cookieID string, chatID string) error {
	query := `INSERT INTO telegram_chats (cookie_id, telegram_chat_id) VALUES ($1, $2)`

	_, err := c.db.ExecContext(context.Background(), query, cookieID, chatID)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("cookieID", cookieID).
			Str("chatID", chatID).
			Msg("Failed to set chatIds")
		return err
	}
	return nil
}

func (c chatIdsRepository) Get(cookieID string) (string, error) {
	query := `SELECT telegram_chat_id FROM telegram_chats WHERE cookie_id=$1`

	telegramChatId := ""

	err := c.db.QueryRowContext(context.Background(), query, cookieID).Scan(&telegramChatId)
	if err != nil {
		zlog.Logger.Error().Err(err).
			Str("cookieID", cookieID).
			Msg("Failed to get telegram_chat_id")
		return "", err
	}

	return telegramChatId, nil
}
