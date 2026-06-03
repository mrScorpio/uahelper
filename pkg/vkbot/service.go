package vkbot

import (
	botgolang "github.com/mail-ru-im/bot-golang"
	"github.com/mrscorpio/uahelper/configs"
)

type VkBot struct {
	B *botgolang.Bot
}

func NewBot(cfg *configs.Config, dontstart bool) (*VkBot, error) {
	if dontstart || !cfg.VkBot {
		return nil, nil
	}

	b, err := botgolang.NewBot(cfg.VkToken)
	if err != nil {
		return nil, err
	}
	mes := b.NewTextMessage(cfg.VkBossId, "логер запущен")
	go mes.Send()

	return &VkBot{B: b}, nil
}
