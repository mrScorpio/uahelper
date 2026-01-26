package tgbot

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func NewBot(token string, dontstart bool) (*bot.Bot, error) {
	if dontstart {
		return nil, nil
	}
	botOpts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(token, botOpts...)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		myChat := update.Message.Chat.ID
		if update.Message.Text == "id" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: myChat,
				Text:   fmt.Sprint("ID твоего чата:", myChat),
			})
		}
	}

}
