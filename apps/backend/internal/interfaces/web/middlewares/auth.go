package middlewares

import (
	"strings"

	"heimdall/backend/internal/application/services"
	"heimdall/backend/internal/constants"
	"heimdall/backend/internal/domain/entities"
	internal_errors "heimdall/backend/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	accessTokenService  services.JWTService
	refreshTokenService services.JWTService
}

func NewAuthMiddleware(ats services.JWTService, rts services.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		accessTokenService:  ats,
		refreshTokenService: rts,
	}
}

func (m *AuthMiddleware) VerifyBearerToken(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")

	if tokenString == "" {
		ctx.Error(internal_errors.NewAuthenticationError("Missing Authentication Token"))
		ctx.Abort()
		return
	}

	if !strings.HasPrefix(tokenString, "Bearer ") {
		ctx.Error(internal_errors.NewAuthenticationError("Invalid token format. Token must start with 'Bearer '"))
		ctx.Abort()
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	claims := m.verifyToken(ctx, m.accessTokenService, tokenString)

	ctx.Set(constants.USER_CLAIMS, claims)
	ctx.Set(constants.USER_ACCESS_TOKEN, tokenString)

	ctx.Next()
}

func (m *AuthMiddleware) VerifyRefreshToken(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")

	claims := m.verifyToken(ctx, m.refreshTokenService, tokenString)

	ctx.Set(constants.USER_CLAIMS, claims)
	ctx.Set(constants.USER_REFRESH_TOKEN, tokenString)

	ctx.Next()
}

func (m *AuthMiddleware) verifyToken(ctx *gin.Context, jwtService services.JWTService, tokenString string) *entities.TokenClaims {
	var claims entities.TokenClaims
	var err error

	claims, err = jwtService.VerifyAndDecode(tokenString)
	if err != nil {
		if _, isInfrastructureError := err.(*internal_errors.InfrastructureError); isInfrastructureError {
			ctx.Error(err)
		} else if _, isValidationError := err.(*internal_errors.ValidationError); isValidationError {
			ctx.Error(internal_errors.NewAuthenticationError("Missing Authentication Token"))
		} else {
			ctx.Error(internal_errors.NewAuthenticationError("The security token is invalid"))
		}
		ctx.Abort()
		return nil
	}

	return &claims
}

func (m *AuthMiddleware) VerifyApiKey(ctx *gin.Context) {

}
