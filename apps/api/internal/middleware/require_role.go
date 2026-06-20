package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.MustGet("role").(string)

		slog.Info("role from context", " ", role)

		for _, r := range roles {
			if role == r {
				slog.Info("valid role")
				c.Next()
				return
			}
		}

		slog.Info("invalid role")

		c.JSON(http.StatusForbidden, types.APIResponse{
			Success: false,
			Message: "Insufficient permissions",
			Code:    constants.Forbidden,
		})
		c.Abort()
	}
}
