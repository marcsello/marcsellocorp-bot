package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/marcsello/marcsellocorp-bot/common"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/memdb"
	"github.com/marcsello/marcsellocorp-bot/utils"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"slices"
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
		msg := fmt.Sprintf("You are %s.", user.Greet())
		if user.IsAdmin() {
			msg += "\nYou are an admin!"
		}
		return ctx.Reply(msg, telebot.ModeDefault)
	}
}

func subscriptionChanging(ctx telebot.Context, state bool) error {

	if len(ctx.Args()) != 1 {
		return ctx.Reply("wrong arguments: /whatever <Channel ID>", telebot.ModeDefault)
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

func cmdMakeChannel(ctx telebot.Context) error {
	user := getUserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("could not get user")
	}

	if len(ctx.Args()) != 1 {
		return ctx.Reply("Usage: /mkchan <Channel name>", telebot.ModeDefault)
	}

	chName := strings.TrimSpace(ctx.Args()[0])

	if !utils.IsValidChannelName(chName) {
		return ctx.Reply("Invalid channel name!", telebot.ModeDefault)
	}

	newChan := db.Channel{
		Name:    chName,
		Creator: user,
	}
	_, err := db.CreateChannel(&newChan)

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ctx.Reply("Channel name already used by current or past channels!\nChannels may not be re-created for security reasons.", telebot.ModeDefault)
		}
		return err
	}

	log.Println("BOT: Channel created: ", ctx.Sender().ID, " -- c:", chName)
	return ctx.Reply("Channel created!", telebot.ModeDefault)

}

func cmdRemoveChannel(ctx telebot.Context) error {
	if len(ctx.Args()) != 1 {
		return ctx.Reply("Usage: /rmchan <Channel name>", telebot.ModeDefault)
	}

	chName := strings.TrimSpace(ctx.Args()[0])

	if !utils.IsValidChannelName(chName) {
		return ctx.Reply("Invalid channel name!", telebot.ModeDefault)
	}

	err := db.DeleteChannelByName(chName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Reply("Channel not found!")
		}
		return err
	}

	log.Println("BOT: Channel deleted: ", ctx.Sender().ID, " -- c:", chName)
	return ctx.Reply("Channel "+chName+" deleted!", telebot.ModeDefault)
}

func cmdListTokens(ctx telebot.Context) error {

	tokens, err := db.GetAllTokens()
	if err != nil {
		return err
	}

	msg := "Currently active tokens:\n"
	for _, token := range tokens {

		var allowedChannelsStr string
		if len(token.AllowedChannels) > 0 {
			allowedChannelsStr = "\n"
			for _, ch := range token.AllowedChannels {
				allowedChannelsStr += "    - " + ch.Name + "\n"
			}
		} else {
			allowedChannelsStr = " <i>NONE!</i>\n"
		}

		lastUsedStr := "Never"
		if token.LastUsed != nil {
			lastUsedStr = token.LastUsed.Format("2006-01-02 15:04:05")
		}

		msg += fmt.Sprintf("- %s\n  <b>created</b>: %s\n  <b>last used</b>: %s\n  <b>allowed channels</b>:%s  <b>capNotify</b>: %s\n  <b>capQuestion</b>: %s\n\n",
			token.Name,
			token.CreatedAt.Format("2006-01-02 15:04:05"),
			lastUsedStr,
			allowedChannelsStr,
			utils.BoolToEmoji(token.CapNotify),
			utils.BoolToEmoji(token.CapQuestion),
		)
	}

	return ctx.Reply(msg, telebot.ModeHTML)
}

func cmdMakeToken(ctx telebot.Context) error {
	const capQuestion = "question"
	const capNotify = "notify"
	validCaps := []string{capQuestion, capNotify}
	if len(ctx.Args()) != 3 {
		return ctx.Reply("Usage: /mktoken <Token name> <Allowed channels comma separated> <Capabilities comma separated>\nValid capabilities: "+strings.Join(validCaps, ", "), telebot.ModeDefault)
	}

	tName := strings.TrimSpace(ctx.Args()[0])

	if !utils.IsValidTokenName(tName) {
		return ctx.Reply("Invalid token name!", telebot.ModeDefault)
	}

	channels := strings.Split(ctx.Args()[1], ",")
	for _, ch := range channels {
		if !utils.IsValidChannelName(ch) {
			return ctx.Reply("Invalid channel name: "+ch+"!", telebot.ModeDefault)
		}
	}

	caps := strings.SplitN(ctx.Args()[2], ",", len(validCaps)+1)
	for _, c := range caps {
		if !slices.Contains(validCaps, c) {
			return ctx.Reply("Invalid capability: "+c+"!", telebot.ModeDefault)
		}
	}
	if len(caps) == 0 {
		return ctx.Reply("Please set at least one capability!", telebot.ModeDefault)
	}

	newTokenStr, err := utils.GenerateRandomString(48)
	if err != nil {
		return err
	}

	newToken := db.Token{
		Name:        tName,
		LastUsed:    nil,
		TokenHash:   utils.TokenHash(newTokenStr),
		CapNotify:   slices.Contains(caps, capNotify),
		CapQuestion: slices.Contains(caps, capQuestion),
	}

	_, err = db.CreateToken(&newToken, channels)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ctx.Reply("This name is already in use!", telebot.ModeDefault)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Reply("Channel not found!", telebot.ModeDefault)
		}
		return err
	}
	message := fmt.Sprintf("<b>New token created!</b>\n<b>Name:</b> %s\n<b>token:</b><pre>%s</pre>\n\n<i>Keep this token a secret, delete this message if possible!</i>", tName, newTokenStr)

	log.Println("BOT: Token created: ", ctx.Sender().ID, " -- t:", tName)
	return ctx.Reply(message, telebot.ModeHTML)
}

func cmdRemoveToken(ctx telebot.Context) error {
	if len(ctx.Args()) != 1 {
		return ctx.Reply("Usage: /rmtoken <Token name>", telebot.ModeDefault)
	}

	tName := strings.TrimSpace(ctx.Args()[0])

	if !utils.IsValidTokenName(tName) {
		return ctx.Reply("Invalid token name!", telebot.ModeDefault)
	}

	err := db.DeleteTokenByName(tName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Reply("Token not found: " + tName + "!")
		}
		return err
	}

	log.Println("BOT: Token deleted: ", ctx.Sender().ID, " -- t:", tName)
	return ctx.Reply("Token "+tName+" deleted!", telebot.ModeDefault)

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

	// Get the original label, or data if not set...
	answerLabel := ""
	for _, op := range questionData.Options {
		if op.Data == cd.Data {
			answerLabel = op.Label
			if answerLabel == "" {
				answerLabel = op.Data
			}
			break
		}
	}

	replyMsg := fmt.Sprintf("Answered by %s:\n\n%s", username, answerLabel)

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
	adminOnly := bot.Group()
	adminOnly.Use(privateOnlyMiddleware)
	adminOnly.Use(knownSenderOnlyMiddleware)
	adminOnly.Use(adminOnlyMiddleware)
	adminOnly.Handle("/mkchan", cmdMakeChannel)
	adminOnly.Handle("/rmchan", cmdRemoveChannel)
	adminOnly.Handle("/tokens", cmdListTokens)
	adminOnly.Handle("/mktoken", cmdMakeToken)
	adminOnly.Handle("/rmtoken", cmdRemoveToken)

	bot.Handle(telebot.OnCallback, handleCallback)
}
