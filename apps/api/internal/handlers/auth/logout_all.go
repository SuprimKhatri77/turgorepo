package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
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
		utils.ClearAuthCookies(c, cfg)

		rlog.Info(c, "logged out on all devices", "user_id", userIDFromContext)
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out on all devices",
		})
	}
}

// revokeUserTokens rejects every access token the user has been issued so far.
// It runs after the sessions are revoked, so a refresh can't issue a token after
// the cutoff. A failure is logged rather than returned: the sessions are already
// revoked, so old access tokens can't be refreshed and expire within minutes.
func revokeUserTokens(c *gin.Context, revoker session.TokenRevoker, cfg *config.Config, userID string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), revokeTimeout)
	defer cancel()

	if err := revoker.RevokeTokensBefore(ctx, userID, time.Now(), cfg.AccessTokenTTL); err != nil {
		rlog.Error(c, "failed to revoke access tokens", err, "user_id", userID)
	}
}
