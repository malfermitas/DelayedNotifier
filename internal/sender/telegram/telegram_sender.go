package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/zlog"
)

type TelegramSender struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramSender(token string) (*TelegramSender, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		zlog.Logger.Err(err)
		return nil, fmt.Errorf("cannot create bot: %v", err)
	}

	return &TelegramSender{bot: bot}, nil
}

func (ts *TelegramSender) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := ts.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("unable to send the message: %v", err)
	}
	return nil
}
