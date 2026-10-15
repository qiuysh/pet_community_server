package utils

import (
	"net/http"

	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// RespondSuccess sends a successful response
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, types.SuccessResponse(data))
}

// RespondError sends an error response
func RespondError(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, types.ErrorResponse(code, message))
}

// RespondValidationError sends a validation error response
func RespondValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, types.ValidationErrorResponse(errors))
}

// getHTTPStatus maps error codes to HTTP status codes
func getHTTPStatus(code int) int {
	switch code {
	case types.CodeUnauthorized:
		return http.StatusUnauthorized
	case types.CodeForbidden:
		return http.StatusForbidden
	case types.CodeNotFound:
		return http.StatusNotFound
	case types.CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case types.CodeRequestTimeout:
		return http.StatusRequestTimeout
	case types.CodeTooManyRequests:
		return http.StatusTooManyRequests
	case types.CodeSystemError, types.CodeDatabaseError, types.CodeCacheError:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
