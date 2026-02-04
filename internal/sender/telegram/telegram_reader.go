package telegram

import (
	"DelayedNotifier/internal/shared"
	"context"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/zlog"
)

type TelegramReader struct {
	bot            *tgbotapi.BotAPI
	chatDataReader shared.TelegramChatIDReader
}

func NewTelegramReader(chatDataReader shared.TelegramChatIDReader, token string) (*TelegramReader, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		zlog.Logger.Err(err)
		return nil, err
	}

	return &TelegramReader{bot: bot, chatDataReader: chatDataReader}, nil
}

func (t *TelegramReader) Run() {
	zlog.Logger.Info().Msg("Telegram reader running")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				t.chatDataReader.ReadChatData(context.Background(), update.Message.Chat.ID, update.Message.CommandArguments())
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi! Now I will send you any incoming notifications.")
				_, err := t.bot.Send(msg)
				if err != nil {
					zlog.Logger.Error().Err(err).Msg("Message send failed")
				}
			}
		}
	}
}
