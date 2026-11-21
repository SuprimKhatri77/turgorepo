package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

// Logout ends this device's session, including its access token, which would
// otherwise keep working until it expires. The user's other sessions stay.
func (h *Handler) Logout(c *gin.Context) {
	sessionID, userID, found := h.sessionFromCookies(c)
	session.ClearCookies(c, h.Config)

	if !found {
		rlog.Info(c, "logout without a valid session token")
		response.OK(c, http.StatusOK, "Logged out successfully", nil)
		return
	}

	// Best-effort revoke: cookies are already cleared, so the client is logged out.
	// Log and still return 200 — failing logout after clearing cookies would confuse the user.
	if familyID, err := utils.ConvertToUUID(sessionID); err != nil {
		rlog.Warn(c, "invalid session id on logout", "error", err)
	} else if err := h.Store.RevokeSession(c.Request.Context(), familyID); err != nil {
		rlog.Error(c, "failed to revoke session on logout", err)
	}
	h.revokeSessionTokens(c, sessionID)

	rlog.Info(c, "user logged out", "user_id", userID, "session_id", sessionID)
	response.OK(c, http.StatusOK, "Logged out successfully", nil)
}
