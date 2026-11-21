package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
)

// Recovery returns a middleware that recovers from panics and responds with 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				rlog.Error(c, "panic recovered", fmt.Errorf("%v", err), "stack", string(debug.Stack()))
				response.Internal(c)
				c.Abort()
			}
		}()
		c.Next()
	}
}
