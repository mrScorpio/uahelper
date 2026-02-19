package tgbot

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mrscorpio/uahelper/configs"
)

type TgBot struct {
	B   *bot.Bot
	cfg *configs.Config
	ctx context.Context
}

func NewBot(ctx context.Context, cfg *configs.Config, dontstart bool) (*TgBot, error) {
	if dontstart || !cfg.Bot {
		return nil, nil
	}
	botOpts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(cfg.BotToken, botOpts...)
	if err != nil {
		return nil, err
	}
	go b.Start(ctx)
	return &TgBot{b, cfg, ctx}, nil
}

func (mybot *TgBot) SendTxt(txt string) error {
	_, err := mybot.B.SendMessage(mybot.ctx, &bot.SendMessageParams{ChatID: mybot.cfg.BotChat, Text: txt})
	if err != nil {
		return err
	}
	return nil
}

func (mybot *TgBot) SendArh(buf *bytes.Buffer, filename string) error {
	if buf == nil {
		return fmt.Errorf("buf is nil, nothing to send by tgbot")
	}
	prms := &bot.SendDocumentParams{
		ChatID:   mybot.cfg.BotChat,
		Document: &models.InputFileUpload{Filename: filename, Data: bytes.NewReader(buf.Bytes())},
		Caption:  "прога для просмотра: https://disk.yandex.ru/d/TXKDwhai1GHSbw",
	}
	_, err := mybot.B.SendDocument(mybot.ctx, prms)

	if err != nil {
		return err
	}
	return nil
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
