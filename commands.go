package main

import (
	"fmt"
	"gopkg.in/telebot.v3"
)

func cmdStart(ctx telebot.Context) error {
	return ctx.Send("Hi there!", telebot.ModeDefault)
}

func cmdId(ctx telebot.Context) error {

	text := fmt.Sprintf("The ID of this chat: %d\nType: %s\n\nID of sender: %d",
		ctx.Chat().ID,
		ctx.Chat().Type,
		ctx.Sender().ID,
	)

	return ctx.Send(text, telebot.ModeDefault)
}
