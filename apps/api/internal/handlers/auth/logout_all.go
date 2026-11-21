package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

// LogoutAll ends every one of the user's sessions, including this one, and
// rejects their access tokens until those expire. It needs RequireAuth.
func LogoutAll(queries repository.AuthRepository, cfg *config.Config, revoker session.TokenRevoker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDFromContext := c.MustGet(constants.UserIDKey).(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			rlog.Error(c, "invalid user_id in context", err, "user_id", userIDFromContext)
			c.JSON(http.StatusBadRequest, types.APIResponse{
				Success: false,
				Message: "Invalid user ID format",
				Code:    constants.ValidationFailed,
			})
			return
		}

		// Nothing has been revoked if this fails, so the user stays logged in and
		// is told, rather than believing their other devices are logged out.
		if err := queries.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
			rlog.Error(c, "failed to revoke all sessions", err, "user_id", userIDFromContext)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to log out of all devices",
				Code:    constants.InternalServerError,
			})
			return
		}

		revokeUserTokens(c, revoker, cfg, userIDFromContext)
		session.ClearCookies(c, cfg)

		rlog.Info(c, "logged out on all devices", "user_id", userIDFromContext)
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out on all devices",
		})
	}
}
