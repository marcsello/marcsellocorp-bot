package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// handleInternalError create a 500 response for error
func handleInternalError(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// handleUserError create a 400 response for error
func handleUserError(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
}

// handleConflict create a 409 response for error
func handleConflict(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"reason": err.Error()})
}
