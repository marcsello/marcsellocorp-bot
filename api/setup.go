package api

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/MikeTTh/env"
)

func InitApi() (func(), error) {

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
