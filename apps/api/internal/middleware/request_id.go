package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

const RequestIDHeader = constants.RequestIDHeader
const RequestIDKey = constants.RequestIDKey

// RequestID ensures every request has an ID: reuse the incoming header or mint one.
// The value is stored on the Gin context and echoed back on the response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}
