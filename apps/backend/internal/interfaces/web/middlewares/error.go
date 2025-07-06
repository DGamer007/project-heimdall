package middlewares

import (
	"fmt"
	"log"
	"net/http"

	"heimdall/backend/internal/application/dto"
	internal_errors "heimdall/backend/pkg/errors"

	"github.com/gin-gonic/gin"
)

func HandleError(ctx *gin.Context) {
	ctx.Next()

	if len(ctx.Errors) > 0 {
		err := ctx.Errors.Last().Err

		switch domainErr := err.(type) {

		case *internal_errors.ValidationError:
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success: false,
					Message: domainErr.Message,
				},
				Errors: convertDetailsToStringMap(domainErr.Details),
			})

		case *internal_errors.AuthenticationError:
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success: false,
					Message: domainErr.Message,
				},
				Errors: nil,
			})

		case *internal_errors.DuplicateResourceError:
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success: false,
					Message: domainErr.Message,
				},
				Errors: convertDetailsToStringMap(domainErr.Details),
			})

		case *internal_errors.NotFoundError:
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success: false,
					Message: domainErr.Message,
				},
				Errors: convertDetailsToStringMap(domainErr.Details),
			})

		case *internal_errors.InfrastructureError:
			log.Print(domainErr.OriginalError)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success: false,
					Message: "An internal error occurred",
				},
				Errors: nil,
			})

		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				BaseResponse: dto.BaseResponse{
					Success: false,
					Message: "An unexpected error occurred",
				},
				Errors: nil,
			})
		}

		ctx.Abort()
	}
}

func convertDetailsToStringMap(details map[string]any) map[string]string {
	if details == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range details {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}
