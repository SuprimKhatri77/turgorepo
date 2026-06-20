package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

// Recovery returns a middleware that recovers from panics and responds with 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, types.APIResponse{
					Success: false,
					Message: "Failed to process request",
					Code:    constants.InternalServerError,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
