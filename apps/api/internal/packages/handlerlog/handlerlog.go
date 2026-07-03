package handlerlog

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func attrs(c *gin.Context, extra ...any) []any {
	a := []any{
		"path", c.FullPath(),
		"method", c.Request.Method,
		"ip", c.ClientIP(),
	}
	if v, ok := c.Get("userID"); ok {
		if s, ok := v.(string); ok && s != "" {
			a = append(a, "actor_id", s)
		}
	}
	return append(a, extra...)
}

func Error(c *gin.Context, msg string, err error, extra ...any) {
	a := attrs(c, extra...)
	a = append(a, "error", err)
	slog.Error(msg, a...)
}

func Warn(c *gin.Context, msg string, extra ...any) {
	slog.Warn(msg, attrs(c, extra...)...)
}

func Info(c *gin.Context, msg string, extra ...any) {
	slog.Info(msg, attrs(c, extra...)...)
}
