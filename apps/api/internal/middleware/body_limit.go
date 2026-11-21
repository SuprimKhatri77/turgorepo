package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
)

// BodyLimit rejects request bodies over limit bytes. Declared sizes are refused
// before reading anything, and MaxBytesReader catches bodies that lie about it.
func BodyLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			response.Abort(c, http.StatusRequestEntityTooLarge, constants.PayloadTooLarge, "Request body is too large")
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
