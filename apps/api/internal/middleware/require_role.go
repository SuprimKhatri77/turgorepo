package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.MustGet(constants.RoleKey).(string)

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		rlog.Warn(c, "insufficient permissions", "role", role)

		c.JSON(http.StatusForbidden, types.APIResponse{
			Success: false,
			Message: "Insufficient permissions",
			Code:    constants.Forbidden,
		})
		c.Abort()
	}
}
