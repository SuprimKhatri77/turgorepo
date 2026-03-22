package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/backend/internal/errors"
)

// Recovery returns a middleware that recovers from panics and responds with 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ResponseMessage(errors.ErrInternal)})
				c.Abort()
			}
		}()
		c.Next()
	}
}
