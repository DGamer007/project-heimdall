package dto

type LoginWithPasswordPayload struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required,min=8"`
}

type TokenPairResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type LoginWithPasswordResponse struct {
	BaseResponse
	Data TokenPairResponse `json:"data"`
}

type RegisterWithPasswordPayload struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	UserName  string `json:"userName" binding:"required"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

type RegisterWithPasswordResponse struct {
	BaseResponse
	Data TokenPairResponse `json:"data"`
}

type RefreshTokenPairResponse struct {
	BaseResponse
	Data TokenPairResponse `json:"data"`
}
