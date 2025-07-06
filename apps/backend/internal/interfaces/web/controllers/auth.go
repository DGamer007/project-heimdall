package controllers

import (
	"net/http"

	"heimdall/backend/internal/application/dto"
	"heimdall/backend/internal/application/services"
	"heimdall/backend/internal/domain/entities"
	web_helpers "heimdall/backend/internal/interfaces/web/helpers"
	internal_errors "heimdall/backend/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service     *services.AuthService
	accessToken services.JWTService
}

func NewAuthController(service *services.AuthService, ats services.JWTService) *AuthController {
	return &AuthController{
		service:     service,
		accessToken: ats,
	}
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	claims, ok := web_helpers.GetClaims(ctx)
	if !ok {
		ctx.Error(internal_errors.NewAuthenticationError("Authentication required"))
		return
	}

	accessToken := ctx.Request.Header.Get("X-Access-Token")
	if accessToken == "" {
		ctx.Error(internal_errors.NewAuthenticationError("Access Token is missing"))
		return
	}

	accessTokenClaims, err := c.accessToken.Decode(accessToken)
	if err != nil {
		ctx.Error(internal_errors.NewAuthenticationError("Invalid Access Token"))
		return
	}

	refreshToken, _ := web_helpers.GetRefreshToken(ctx)

	var tokenPair entities.TokenPair
	tokenPair, err = c.service.RefreshTokenPair(claims.Subject, accessTokenClaims.JTI, refreshToken)
	if err != nil {
		ctx.Error(err)
		return
	}

	response := dto.RefreshTokenPairResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "Successfully generated new token pair",
		},
		Data: dto.TokenPairResponse{
			AccessToken:  tokenPair.AccessToken.SignedToken,
			RefreshToken: tokenPair.RefreshToken.SignedToken,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	claims, ok := web_helpers.GetClaims(ctx)
	if !ok {
		ctx.Error(internal_errors.NewAuthenticationError("Authentication required"))
		return
	}

	_, err := c.service.Logout(claims.Subject, claims.JTI)
	if err != nil {
		ctx.Error(err)
		return
	}

	response := dto.BaseResponse{
		Success: true,
		Message: "Successfully logged out",
	}

	ctx.JSON(http.StatusOK, response)
}
