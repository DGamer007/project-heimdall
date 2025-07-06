package controllers

import (
	"net/http"

	"heimdall/backend/internal/application/dto"
	"heimdall/backend/internal/application/services"
	internal_string "heimdall/backend/pkg/string"
	internal_struct "heimdall/backend/pkg/struct"

	"github.com/gin-gonic/gin"
)

type LocalAuthController struct {
	service *services.LocalAuthService
}

func NewLocalAuthController(service *services.LocalAuthService) *LocalAuthController {
	return &LocalAuthController{
		service: service,
	}
}

func (c *LocalAuthController) Login(ctx *gin.Context) {
	var body dto.LoginWithPasswordPayload
	if err := internal_struct.ValidateAndBind(ctx, &body); err != nil {
		ctx.Error(err)
		return
	}

	// Comprehensive validation for SQL injection and XSS before sanitization
	if err := internal_string.ValidateUserInput(body.Identifier, "identifier"); err != nil {
		ctx.Error(err)
		return
	}

	// Sanitize user inputs to prevent XSS attacks (after validation)
	body.Identifier = internal_string.SanitizeStrict(body.Identifier)

	tokenPair, err := c.service.LoginWithPassword(body)
	if err != nil {
		ctx.Error(err)
		return
	}

	response := dto.LoginWithPasswordResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "Successfully logged in",
		},
		Data: dto.TokenPairResponse{
			AccessToken:  tokenPair.AccessToken.SignedToken,
			RefreshToken: tokenPair.RefreshToken.SignedToken,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *LocalAuthController) Register(ctx *gin.Context) {
	var body dto.RegisterWithPasswordPayload
	if err := internal_struct.ValidateAndBind(ctx, &body); err != nil {
		ctx.Error(err)
		return
	}

	// Comprehensive validation for SQL injection and XSS before sanitization
	if err := internal_string.ValidateUserInput(body.FirstName, "firstName"); err != nil {
		ctx.Error(err)
		return
	}

	if err := internal_string.ValidateUserInput(body.LastName, "lastName"); err != nil {
		ctx.Error(err)
		return
	}

	if err := internal_string.ValidateUserInput(body.UserName, "userName"); err != nil {
		ctx.Error(err)
		return
	}

	// Sanitize user inputs to prevent XSS attacks (after validation)
	body.FirstName = internal_string.SanitizeStrict(body.FirstName)
	body.LastName = internal_string.SanitizeStrict(body.LastName)
	body.UserName = internal_string.SanitizeStrict(body.UserName)

	tokenPair, err := c.service.RegisterWithPassword(body)
	if err != nil {
		ctx.Error(err)
		return
	}

	response := dto.RegisterWithPasswordResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: "Successfully registered new account",
		},
		Data: dto.TokenPairResponse{
			AccessToken:  tokenPair.AccessToken.SignedToken,
			RefreshToken: tokenPair.RefreshToken.SignedToken,
		},
	}

	ctx.JSON(http.StatusCreated, response)
}
