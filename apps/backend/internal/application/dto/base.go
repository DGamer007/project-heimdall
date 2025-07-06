package dto

type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	BaseResponse
	Errors map[string]string `json:"errors,omitempty"`
}
