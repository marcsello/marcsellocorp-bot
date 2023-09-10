package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/marcsello/marcsellocorp-bot/common"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/memdb"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"strings"
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

	return ctx.Reply(text, telebot.ModeDefault)
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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Reply("channel not found", telebot.ModeDefault)
		}
		return err
	}

	var changed bool
	changed, err = db.ChangeSubscription(ctx.Chat().ID, ch.ID, state)
	if err != nil {
		return err
	}

	var msg string
	if changed {
		if state {
			msg = "Successfully subscribed to " + chName
		} else {
			msg = "Successfully unsubscribed from " + chName
		}
	} else {
		if state {
			msg = "Already subscribed to " + chName
		} else {
			msg = "Not subscribed to " + chName
		}
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

func handleCallback(ctx telebot.Context) error {
	q := ctx.Callback()

	parts := strings.SplitN(strings.TrimSpace(q.Data), "|", 2)

	if len(parts) != 2 { // first is the unique part
		return fmt.Errorf("invalid or no data passed")
	}

	unique := strings.TrimSpace(parts[0])
	data := strings.TrimSpace(parts[1])

	log.Println("BOT: New callback: ", ctx.Sender().ID, " -- u: ", unique, " -- d: ", data)

	if unique != common.CallbackIDQuestion {
		return nil
	}

	var cd common.CallbackData
	err := json.Unmarshal([]byte(data), &cd)
	if err != nil {
		return err
	}

	var questionData *memdb.QuestionData
	questionData, err = memdb.AnswerQuestion(context.TODO(), cd.RandomID, ctx.Sender().ID, cd.Data)
	if err != nil {
		return err
	}

	// Update sent messages

	username := "Anon"
	if ctx.Sender().Username != "" {
		username = "@" + ctx.Sender().Username
	} else if ctx.Sender().FirstName != "" {
		username = ctx.Sender().FirstName
	}

	answerLabel := ctx.Text()

	replyMsg := fmt.Sprintf("Answered by %s:\n%s", username, answerLabel)

	for _, sMsg := range questionData.RelatedMessages {
		var msg *telebot.Message
		msg, err = telegramBot.EditReplyMarkup(sMsg, nil) // remove buttons
		if err != nil {
			return err
		}
		_, err = telegramBot.Reply(msg, replyMsg, telebot.ModeDefault)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupHandlers(bot *telebot.Bot) {
	bot.Handle("/start", cmdStart)
	bot.Handle("/id", cmdId)
	bot.Handle("/whoami", cmdWhoami)

	privateAuthOnly := bot.Group()
	privateAuthOnly.Use(privateOnlyMiddleware)
	privateAuthOnly.Use(knownSenderOnlyMiddleware)
	privateAuthOnly.Handle("/subscribe", cmdSubscribe)
	privateAuthOnly.Handle("/unsubscribe", cmdUnsubscribe)
	privateAuthOnly.Handle("/list", cmdList)

	bot.Handle(telebot.OnCallback, handleCallback)
}
