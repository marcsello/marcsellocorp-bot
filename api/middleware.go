package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/marcsellocorp-bot/db"
	"github.com/marcsello/marcsellocorp-bot/utils"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func getTokenFromContext(ctx *gin.Context) *db.Token {
	tInt, ok := ctx.Get("token")
	if !ok {
		return nil
	}

	if tInt == nil {
		return nil
	}

	t, ok := tInt.(db.Token)

	if !ok {
		return nil
	}

	return &t
}

func requireValidTokenMiddleware(ctx *gin.Context) {

	key, ok := parseAuthHeader(ctx, "Bearer")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := db.GetAndUpdateTokenByHash(utils.TokenHash(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		} else {
			handleInternalError(ctx, err)
			ctx.Abort()
		}
		return
	}
	if token == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set("token", token)
}

func parseAuthHeader(ctx *gin.Context, type_ string) (string, bool) {
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", false
	}

	if parts[0] != type_ {
		return "", false
	}

	return parts[1], true
}
