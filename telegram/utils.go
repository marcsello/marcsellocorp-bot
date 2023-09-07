package telegram

import (
	"github.com/marcsello/marcsellocorp-bot/db"
	"gopkg.in/telebot.v3"
)

func getUserFromContext(ctx telebot.Context) *db.User {

	uInt := ctx.Get("user")

	if uInt == nil {
		return nil
	}

	u, ok := uInt.(db.User)
	
	if !ok {
		return nil
	}

	return &u
}
