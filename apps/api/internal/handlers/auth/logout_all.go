package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

// LogoutAll ends every one of the user's sessions, including this one, and
// rejects their access tokens until those expire. It needs RequireAuth.
func (h *Handler) LogoutAll(c *gin.Context) {
	userIDFromContext, userID, ok := currentUserID(c)
	if !ok {
		return
	}

	// Nothing has been revoked if this fails, so the user stays logged in and
	// is told, rather than believing their other devices are logged out.
	if err := h.Store.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
		rlog.Error(c, "failed to revoke all sessions", err, "user_id", userIDFromContext)
		response.Internal(c)
		return
	}

	h.revokeUserTokens(c, userIDFromContext)
	session.ClearCookies(c, h.Config)

	rlog.Info(c, "logged out on all devices", "user_id", userIDFromContext)
	response.OK(c, http.StatusOK, "Logged out on all devices", nil)
}
