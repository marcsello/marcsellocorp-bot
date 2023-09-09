package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/marcsellocorp-bot/common"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/memdb"
	"github.com/marcsello/marcsellocorp-bot/telegram"
	"gopkg.in/telebot.v3"
	"net/http"
)

func handleNotify(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token != nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}

	if !token.CapNotify {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "capability disallowed"})
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

		_, err = telegram.GetBot().Send(&telebot.User{ID: sub.ID}, msg, telebot.ModeDefault)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		delivered = true
	}

	resp := NotifyResponse{DeliveredToAnyone: delivered}

	ctx.JSON(http.StatusOK, resp)
}

func handleNewQuestion(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token != nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}
	if !token.CapQuestion {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "capability disallowed"})
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

	var newQuestionTx memdb.NewQuestionTx
	newQuestionTx, err = memdb.BeginNewQuestion(ctx, token.ID)
	if err != nil {
		handleInternalError(ctx, err)
		return
	}

	// compile message
	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, len(req.Options))
	for i, op := range req.Options {
		var data []byte
		data, err = json.Marshal(common.CallbackData{
			RandomID: newQuestionTx.RandomID(),
			Data:     op.Data,
		})
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		rows[i] = markup.Row(markup.Data(op.Label, common.CallbackIDQuestion, string(data)))
	}
	markup.Inline(rows...)

	msg := fmt.Sprintf("[%s -> %s]\n\n%s", token.Name, targetChannel.Name, req.Text)

	for _, sub := range targetChannel.Subscribers {
		var m *telebot.Message
		m, err = telegram.GetBot().Send(&telebot.User{ID: sub.ID}, msg, telebot.ModeDefault, markup)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		newQuestionTx.AddRelatedMessage(memdb.StoredMessage{MessageID: m.ID, ChatID: m.Chat.ID})
	}

	err = newQuestionTx.Close()
	if err != nil {
		handleInternalError(ctx, err)
		return
	}

	resp := QuestionResponse{
		ID: newQuestionTx.RandomID(),
	}

	ctx.JSON(http.StatusCreated, resp)
}

func handleQuestionAnswer(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token != nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}
	if !token.CapQuestion {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "capability disallowed"})
		return
	}

	id := ctx.Param("id")

	q, err := memdb.GetQuestionData(ctx, id)
	if err != nil {
		handleInternalError(ctx, err)
		return
	}

	if q == nil || q.SourceTokenID != token.ID {
		ctx.Status(http.StatusNotFound)
		return
	}

	resp := QuestionResponse{
		ID:     id,
		Answer: nil,
	}

	// Check if question is answered
	if q.AnswerData != nil && q.AnsweredAt != nil && q.AnswererID != nil {

		// get answerer data from db
		var user *db.User
		user, err = db.GetUserById(*q.AnswererID)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		// fill answerer in response
		resp.Answer = &QuestionAnswer{
			Data:       *q.AnswerData,
			AnsweredAt: *q.AnsweredAt,
			AnsweredBy: UserToUserRepr(*user),
		}
	}

	ctx.JSON(http.StatusOK, resp)
}
