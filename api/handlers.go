package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/marcsellocorp-bot/common"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/memdb"
	"github.com/marcsello/marcsellocorp-bot/telegram"
	"github.com/redis/go-redis/v9"
	"gopkg.in/telebot.v3"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

func handleNotify(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token == nil {
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
	if req.Text == "" {
		handleUserError(ctx, fmt.Errorf("text may not be empty"))
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

	log.Println("API: New notification created: ", token.Name, " -- ch: ", targetChannel.Name)
	ctx.JSON(http.StatusOK, resp)
}

func handleNewQuestion(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token == nil {
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
	if len(req.Options) == 0 {
		handleUserError(ctx, fmt.Errorf("no options provided"))
		return
	}
	if req.Text == "" {
		handleUserError(ctx, fmt.Errorf("text may not be empty"))
		return
	}
	for _, op := range req.Options {
		if op.Data == "" {
			handleUserError(ctx, fmt.Errorf("option data must be defined"))
			return
		}
		if len(op.Data) > 12 {
			/*
				>>> len('{"i":"w08slJSreJu2VzVhyzBZYnGMPj7kfEpk","d":""}')
				47
				>>> 64-47
				17
			*/
			handleUserError(ctx, fmt.Errorf("max size for data is 12 bytes"))
			return
		}
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

		newQuestionTx.AddOption(op.Data, op.Label) // store the original label in redis

		label := op.Label
		if label == "" {
			label = op.Data
		}

		var btnData []byte
		btnData, err = json.Marshal(common.CallbackData{
			RandomID: newQuestionTx.RandomID(),
			Data:     op.Data,
		})
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		rows[i] = markup.Row(markup.Data(label, common.CallbackIDQuestion, string(btnData)))
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

	log.Println("API: New question created: ", token.Name, " -- ch: ", targetChannel.Name, " -- op:", len(req.Options))
	ctx.JSON(http.StatusCreated, resp)
}

func memdbAnswerToApiResponse(id string, q memdb.QuestionData) (QuestionResponse, error) {
	var err error

	resp := QuestionResponse{
		ID:     id,
		Answer: nil,
	}

	// Check if question is answered
	if q.IsAnswered() {

		// get answerer data from db
		var user *db.User
		user, err = db.GetUserById(*q.AnswererID)
		if err != nil {
			return resp, err
		}

		// fill answerer in response
		resp.Answer = &QuestionAnswer{
			Data:       *q.AnswerData,
			AnsweredAt: *q.AnsweredAt,
			AnsweredBy: UserToUserRepr(*user),
		}
	}
	return resp, nil

}

func handleQuestionAnswer(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token == nil {
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
		if errors.Is(err, redis.Nil) {
			ctx.Status(http.StatusNotFound)
			return
		}

		handleInternalError(ctx, err)
		return
	}

	if q == nil || q.SourceTokenID != token.ID {
		ctx.Status(http.StatusNotFound)
		return
	}

	var resp QuestionResponse
	resp, err = memdbAnswerToApiResponse(id, *q)
	if err != nil {
		handleInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func handleQuestionAnswerPolling(ctx *gin.Context) {
	token := getTokenFromContext(ctx)
	if token == nil {
		handleInternalError(ctx, fmt.Errorf("invalid token"))
		return
	}
	if !token.CapQuestion {
		ctx.JSON(http.StatusForbidden, gin.H{"reason": "capability disallowed"})
		return
	}

	id := ctx.Param("id")

	// we have to load the data at least once, to determine if we are allowed to read it
	preCheckQ, err := memdb.GetQuestionData(ctx, id)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ctx.Status(http.StatusNotFound)
			return
		}
		handleInternalError(ctx, err)
		return
	}

	if preCheckQ == nil || preCheckQ.SourceTokenID != token.ID {
		ctx.Status(http.StatusNotFound)
		return
	}

	var resp QuestionResponse

	// well, it looks like it's already answered...
	if preCheckQ.IsAnswered() {
		resp, err = memdbAnswerToApiResponse(id, *preCheckQ)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// we are allowed to read it, and not answered yet, set up waiting meme

	ctx2, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	answeredQuestionChan := make(chan *memdb.QuestionData)
	answeredQuestionErrChan := make(chan error)
	defer close(answeredQuestionChan)
	defer close(answeredQuestionErrChan)

	connectionClosed := atomic.Bool{}
	defer connectionClosed.Store(true)

	go func() {
		q, internalErr := memdb.WaitForAnswer(ctx2, id)
		if connectionClosed.Load() {
			// channels are possibly closed, or will be closed soon, don't send anything on them
			return
		}

		if internalErr != nil {
			answeredQuestionErrChan <- internalErr
			return
		}
		answeredQuestionChan <- q
	}()

	select {
	case q := <-answeredQuestionChan:

		if q == nil { // no answer arrived
			ctx.Status(http.StatusNoContent)
			return
		}

		resp, err = memdbAnswerToApiResponse(id, *q)
		if err != nil {
			handleInternalError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, resp)
		return

	case err = <-answeredQuestionErrChan:
		if errors.Is(err, redis.Nil) {
			ctx.Status(http.StatusNotFound)
			return
		}
		handleInternalError(ctx, err)
		return
	case <-ctx.Writer.CloseNotify():
		// client closed request, ctx2 close is deferred
		return
	case <-ctx.Request.Context().Done():
		// client closed request, ctx2 close is deferred
		return
	}

}
