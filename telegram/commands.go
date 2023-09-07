package telegram

import (
	"errors"
	"fmt"
	"github.com/marcsello/marcsellocorp-bot/db"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
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

func cmdWhoami(ctx telebot.Context) error {

	var err error
	var user *db.User
	user, err = db.GetUserById(ctx.Sender().ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// otherwise just ignore
		user = nil
	}

	if (user == nil) || (!user.IsActive()) {
		return ctx.Reply("Sorry, I don't know you.", telebot.ModeDefault)
	} else {
		msg := fmt.Sprintf("You are %s", user.Greet())
		return ctx.Reply(msg, telebot.ModeDefault)
	}
}

func subscriptionChanging(ctx telebot.Context, state bool) error {

	if len(ctx.Args()) != 1 {
		return ctx.Reply("wrong arguments", telebot.ModeDefault)
	}

	chName := ctx.Args()[0]

	ch, err := db.GetChannelByName(chName)

	err = db.ChangeSubscription(ctx.Chat().ID, ch.ID, state)
	if err != nil {
		// Handle already subscribed/unsubscribed condition
		return err
	}

	var msg string
	if state {
		msg = "Successfully subscribed to " + chName
	} else {
		msg = "Successfully unsubscribed from " + chName
	}

	return ctx.Reply(msg, telebot.ModeDefault)

}

func cmdSubscribe(ctx telebot.Context) error {
	return subscriptionChanging(ctx, true)
}

func cmdUnsubscribe(ctx telebot.Context) error {
	return subscriptionChanging(ctx, false)
}

func cmdList(ctx telebot.Context) error {
	user := getUserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("could not get user")
	}

	channels, err := db.GetAllChannels()
	if err != nil {
		return err
	}

	msg := "Available channels:\n"
	for _, ch := range channels {

		prefix := "-"
		for _, sub := range user.Subscriptions {
			if sub.ID == ch.ID {
				prefix = "+"
				break
			}
		}

		msg += fmt.Sprintf(" %s %s\n", prefix, ch.Name)
	}

	return ctx.Reply(msg, telebot.ModeDefault)

}

func setupCommands(bot *telebot.Bot) {
	bot.Handle("/start", cmdStart)
	bot.Handle("/id", cmdId)
	bot.Handle("/whoami", cmdWhoami)

	privateAuthOnly := bot.Group()
	privateAuthOnly.Use(privateOnlyMiddleware)
	privateAuthOnly.Use(knownSenderOnlyMiddleware)
	privateAuthOnly.Handle("/subscribe", cmdSubscribe)
	privateAuthOnly.Handle("/unsubscribe", cmdUnsubscribe)
	privateAuthOnly.Handle("/list", cmdList)
}
