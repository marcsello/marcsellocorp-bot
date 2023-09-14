package telegram

import (
	"errors"
	"fmt"
	"github.com/marcsello/marcsellocorp-bot/db"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

const insufficentPermissionMessage = "You may not use this command"

func privateOnlyMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {

		if ctx.Chat().Type != telebot.ChatPrivate {
			return ctx.Reply("This command is restricted to private chats!", telebot.ModeDefault)
		}

		return next(ctx)
	}
}

func knownSenderOnlyMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
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
			return ctx.Reply(insufficentPermissionMessage, telebot.ModeDefault)
		} else {
			ctx.Set("user", user)
			return next(ctx)
		}
	}
}

func adminOnlyMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		user := getUserFromContext(ctx)
		if user == nil {
			return fmt.Errorf("could not get user")
		}
		if !user.IsAdmin() {
			return ctx.Reply(insufficentPermissionMessage, telebot.ModeDefault)
		}
		return next(ctx)
	}
}
