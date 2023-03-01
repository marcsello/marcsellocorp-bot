package main

import (
	"gitlab.com/MikeTTh/env"
	"gopkg.in/telebot.v3"
	"log"
)

func main() {
	log.Println("Staring Marcsello Corp. Telegram Bot...")

	bot, err := telebot.NewBot(telebot.Settings{
		Token: env.StringOrPanic("TELEGRAM_TOKEN"),
		Poller: &telebot.Webhook{
			Listen: ":8080",
			Endpoint: &telebot.WebhookEndpoint{
				PublicURL: env.StringOrPanic("PUBLIC_URL"),
			},
		},
	})

	if err != nil {
		panic(err)
	}

	bot.Handle("/id", cmdId)
	bot.Handle("/start", cmdStart)

	log.Println("Everything is ready! Listening for commands!")
	bot.Start()
}
