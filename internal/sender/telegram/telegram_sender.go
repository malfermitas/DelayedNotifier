package telegram

import (
	"context"
	"fmt"
	"strconv"

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

func (ts *TelegramSender) Send(ctx context.Context, to string, text string) error {
	chatID, err := strconv.ParseInt(to, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat id: %w", err)
	}
	msg := tgbotapi.NewMessage(chatID, text)
	_, err = ts.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("unable to send the message: %v", err)
	}
	return nil
}
