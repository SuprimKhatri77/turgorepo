package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
)

// healthPath is called constantly by the load balancer, so it isn't logged.
const healthPath = "/api/v1/health"

// RequestLogger writes one structured line per request, for production log tools.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if c.FullPath() == healthPath {
			return
		}
		rlog.Info(c, "request completed",
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
