package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
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

		response.Abort(c, http.StatusForbidden, constants.Forbidden, "Insufficient permissions")
	}
}
