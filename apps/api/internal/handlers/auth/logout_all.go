package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

// LogoutAll ends every one of the user's sessions, including this one, and
// rejects their access tokens until those expire. It needs RequireAuth.
func LogoutAll(queries Store, cfg *config.Config, revoker session.TokenRevoker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDFromContext := c.MustGet(constants.UserIDKey).(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			rlog.Error(c, "invalid user_id in context", err, "user_id", userIDFromContext)
			response.Error(c, http.StatusBadRequest, constants.ValidationFailed, "Invalid user ID format")
			return
		}

		// Nothing has been revoked if this fails, so the user stays logged in and
		// is told, rather than believing their other devices are logged out.
		if err := queries.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
			rlog.Error(c, "failed to revoke all sessions", err, "user_id", userIDFromContext)
			response.Internal(c)
			return
		}

		revokeUserTokens(c, revoker, cfg, userIDFromContext)
		session.ClearCookies(c, cfg)

		rlog.Info(c, "logged out on all devices", "user_id", userIDFromContext)
		response.OK(c, http.StatusOK, "Logged out on all devices", nil)
	}
}
