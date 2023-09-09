package telegram

import (
	"gitlab.com/MikeTTh/env"
	"gopkg.in/telebot.v3"
)

var telegramBot *telebot.Bot

func InitTelegramBot() (func(), error) {
	var err error
	telegramBot, err = telebot.NewBot(telebot.Settings{
		Token: env.StringOrPanic("TELEGRAM_TOKEN"),
		Poller: &telebot.Webhook{
			Listen: env.String("WEBHOOK_BIND", ":8080"),
			Endpoint: &telebot.WebhookEndpoint{
				PublicURL: env.StringOrPanic("WEBHOOK_PUBLIC_URL"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	setupCommands(telegramBot)

	runFunc := func() {
		telegramBot.Start()
	}
	
	return runFunc, nil
}
