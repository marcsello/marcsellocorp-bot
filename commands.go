package main

import (
	"fmt"
	"gopkg.in/telebot.v3"
)

func cmdStart(ctx telebot.Context) error {
	return ctx.Send("Hi there!", telebot.ModeDefault)
}

func cmdId(ctx telebot.Context) error {

	text := fmt.Sprintf("The ID of this chat: %d", ctx.Sender().ID)

	return ctx.Send(text, telebot.ModeDefault)
}
