package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
)

// RequestLogger writes one structured line per request, for production log tools.
// Requests to skipPaths aren't logged, for endpoints a load balancer calls constantly.
func RequestLogger(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]bool, len(skipPaths))
	for _, path := range skipPaths {
		skip[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if skip[c.FullPath()] {
			return
		}
		rlog.Info(c, "request completed",
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
