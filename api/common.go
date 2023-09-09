package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// handleInternalError create a 500 response for error
func handleInternalError(ctx *gin.Context, err error) {
	log.Println("INTERNAL ERROR HAPPENED: ", err)
	ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// handleUserError create a 400 response for error
func handleUserError(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
}
