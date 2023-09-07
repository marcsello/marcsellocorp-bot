package api

import (
	"crypto/rand"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/marcsellocorp-bot/db"
	"gopkg.in/telebot.v3"
	"math/big"
	"net/http"
)

func handleNotify(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token != nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}

	if !token.NotifyAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "method disallowed"})
		return
	}

	var req NotifyRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		handleUserError(ctx, err)
		return
	}

	var targetChannel *db.Channel = nil

	for _, ch := range token.AllowedChannels {
		if ch.Name == req.Channel {
			targetChannel = ch
			break
		}
	}
	if targetChannel == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"reason": "channel not found or no permission"})
		return
	}

	// fill subscribers basically
	targetChannel, err = db.GetChannelById(targetChannel.ID)

	msg := fmt.Sprintf("[%s -> %s]\n\n%s", token.Name, targetChannel.Name, req.Text)

	delivered := false
	for _, sub := range targetChannel.Subscribers {

		_, err = telegramBot.Send(&telebot.User{ID: sub.ID}, msg, telebot.ModeDefault)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		delivered = true
	}

	resp := NotifyResponse{DeliveredToAnyone: delivered}

	ctx.JSON(http.StatusOK, resp)
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func handleNewQuestion(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token != nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}
	if !token.QuestionAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "method disallowed"})
		return
	}

	var req QuestionRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		handleUserError(ctx, err)
		return
	}

	var targetChannel *db.Channel = nil

	for _, ch := range token.AllowedChannels {
		if ch.Name == req.Channel {
			targetChannel = ch
			break
		}
	}
	if targetChannel == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"reason": "channel not found or no permission"})
		return
	}

	// fill subscribers basically
	targetChannel, err = db.GetChannelById(targetChannel.ID)

	if len(targetChannel.Subscribers) == 0 {
		handleUserError(ctx, fmt.Errorf("no subscribers on this channel"))
		return
	}

	var randomID string
	randomID, err = generateRandomString(64)

	q := &db.PendingQuestion{
		RandomID: randomID,
		Source:   token,
	}
	q, err = db.NewPendingQuestion(q)
	if err != nil {
		handleInternalError(ctx, err)
		return
	}

	msg := "" //TODO
	messageIDs := make([]int, len(targetChannel.Subscribers))
	for i, sub := range targetChannel.Subscribers {

		var m *telebot.Message
		m, err = telegramBot.Send(&telebot.User{ID: sub.ID}, msg, telebot.ModeDefault)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		messageIDs[i] = m.ID
	}

	// TODO: Store message IDs in DB

}

func handleQuestionAnswer(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token != nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}
	if !token.QuestionAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "method disallowed"})
		return
	}

	// TODO

}
