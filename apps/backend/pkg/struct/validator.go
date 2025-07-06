package internal_struct

import (
	"encoding/json"
	"fmt"
	"strings"

	internal_errors "heimdall/backend/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(errs validator.ValidationErrors) map[string]any {
	errors := make(map[string]any)

	for _, err := range errs {
		field := strings.ToLower(err.Field())

		switch err.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errors[field] = "Must be a valid email address"
		case "min":
			errors[field] = fmt.Sprintf("%s must be at least %s characters long", field, err.Param())
		case "max":
			errors[field] = fmt.Sprintf("%s must not exceed %s characters", field, err.Param())
		case "url":
			errors[field] = "Must be a valid url"
		default:
			errors[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return errors
}

func Validate[T any](obj *T, messageArgs ...string) error {
	msg := getErrorMessage(messageArgs...)
	validate := validator.New()
	if err := validate.Struct(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return internal_errors.NewValidationError(msg, FormatValidationErrors(validationErrors))
		}

		return err
	}

	return nil
}

func ValidateAndBind[T any](ctx *gin.Context, obj *T, messageArgs ...string) error {
	msg := getErrorMessage(messageArgs...)

	if err := ctx.ShouldBind(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return internal_errors.NewValidationError(msg, FormatValidationErrors(validationErrors))
		} else if unmarshalError, ok := err.(*json.UnmarshalTypeError); ok {
			return internal_errors.NewValidationError(msg, map[string]any{
				unmarshalError.Field: "Field '" + unmarshalError.Field + "' must be a " + unmarshalError.Type.String() + ", but received a " + string(unmarshalError.Value),
			})
		} else {
			return internal_errors.NewValidationError(msg, map[string]any{
				"error": err.Error(),
			})
		}
	}

	return nil
}

func getErrorMessage(messageArgs ...string) string {
	if len(messageArgs) == 0 || messageArgs[0] == "" {
		return "Validation failed"
	}

	return messageArgs[0]
}
