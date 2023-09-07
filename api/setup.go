package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/MikeTTh/env"
	"gopkg.in/telebot.v3"
)

var telegramBot *telebot.Bot // global meme

func InitApi(bot *telebot.Bot) (func(), error) {
	telegramBot = bot

	router := gin.New()
	router.Use(requireValidTokenMiddleware)
	// this is RPC style instead of REST style
	router.POST("/notify", handleNotify)
	router.POST("/question", handleNewQuestion)
	router.GET("/question/:id", handleQuestionAnswer)

	runFunc := func() {
		err := router.Run(env.String("API_BIND", ":8081"))
		if err != nil {
			panic(err)
		}
	}

	return runFunc, nil
}
