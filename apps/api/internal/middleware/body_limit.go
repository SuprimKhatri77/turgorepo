package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

// BodyLimit rejects request bodies over limit bytes. Declared sizes are refused
// before reading anything, and MaxBytesReader catches bodies that lie about it.
func BodyLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, types.APIResponse{
				Success: false,
				Message: "Request body is too large",
				Code:    constants.PayloadTooLarge,
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
