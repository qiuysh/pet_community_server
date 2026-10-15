// server/middleware/error.go
package middleware

import (
	"log"
	"net/http"

	"pet-community/types"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Error: %v", err)

			code := types.CodeSystemError
			message := "System error"

			if e, ok := err.Err.(types.ApiError); ok {
				code = e.Code
				message = e.Message
			} else {
				message = err.Error()
			}

			c.JSON(http.StatusOK, types.ErrorResponse(code, message))
		}
	}
}
