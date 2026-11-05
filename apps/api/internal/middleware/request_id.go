package middleware

import (
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

const RequestIDHeader = constants.RequestIDHeader
const RequestIDKey = constants.RequestIDKey

const maxRequestIDLength = 128

// RequestID ensures every request has an ID: reuse the incoming header or mint one.
// The value is stored on the Gin context and echoed back on the response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !validRequestID(id) {
			id = uuid.NewString()
		}

		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// validRequestID limits what a client can put in our logs through this header.
func validRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}
