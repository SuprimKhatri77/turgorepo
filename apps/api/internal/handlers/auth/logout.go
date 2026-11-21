package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

// Logout ends this device's session, including its access token, which would
// otherwise keep working until it expires. The user's other sessions stay.
func Logout(queries repository.AuthRepository, cfg *config.Config, revoker session.TokenRevoker) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, userID, found := sessionFromCookies(c, cfg)
		utils.ClearAuthCookies(c, cfg)

		if !found {
			rlog.Info(c, "logout without a valid session token")
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Logged out successfully",
			})
			return
		}

		// Best-effort revoke: cookies are already cleared, so the client is logged out.
		// Log and still return 200 — failing logout after clearing cookies would confuse the user.
		if familyID, err := utils.ConvertToUUID(sessionID); err != nil {
			rlog.Warn(c, "invalid session id on logout", "error", err)
		} else if err := queries.RevokeSession(c.Request.Context(), familyID); err != nil {
			rlog.Error(c, "failed to revoke session on logout", err)
		}
		revokeSessionTokens(c, revoker, cfg, sessionID)

		rlog.Info(c, "user logged out", "user_id", userID, "session_id", sessionID)
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out successfully",
		})
	}
}
