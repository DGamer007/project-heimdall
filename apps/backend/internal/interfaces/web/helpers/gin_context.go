package web_helpers

import (
	"heimdall/backend/internal/constants"
	"heimdall/backend/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

func GetClaims(ctx *gin.Context) (*entities.TokenClaims, bool) {
	claims, exists := ctx.Get(constants.USER_CLAIMS)
	if !exists {
		return nil, false
	}

	tokenClaims, ok := claims.(*entities.TokenClaims)
	return tokenClaims, ok
}

func GetAccessToken(ctx *gin.Context) (token string, ok bool) {
	token = ctx.GetString(constants.USER_ACCESS_TOKEN)
	return token, token != ""
}

func GetRefreshToken(ctx *gin.Context) (token string, ok bool) {
	token = ctx.GetString(constants.USER_REFRESH_TOKEN)
	return token, token != ""
}
