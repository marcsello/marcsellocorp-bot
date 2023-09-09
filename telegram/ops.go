package telegram

import "gopkg.in/telebot.v3"

func GetBot() *telebot.Bot {
	return telegramBot
}
