package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

// revocationCheckTimeout keeps a slow Redis from delaying every authenticated request.
const revocationCheckTimeout = 200 * time.Millisecond

func RequireAuth(cfg *config.Config, revoker session.TokenRevoker) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie(session.AccessTokenCookie)
		if err != nil {
			rlog.Warn(c, "missing access token")

			response.Abort(c, http.StatusUnauthorized, constants.TokenNotProvided, "Missing access token")
			return
		}

		claims, err := session.ParseAccess(accessToken, cfg.JWTAccessSecret)
		if err != nil {
			rlog.Warn(c, "invalid access token", "error", err)

			response.Abort(c, http.StatusUnauthorized, constants.TokenInvalid, "Invalid access token")
			return
		}

		if !constants.IsValidRole(claims.Role) {
			rlog.Warn(c, "invalid role in claims", "user_id", claims.UserID, "role", claims.Role)

			response.Abort(c, http.StatusUnauthorized, constants.Unauthorized, "Invalid token claims")
			return
		}

		if tokenRevoked(c, revoker, claims) {
			rlog.Warn(c, "revoked access token", "user_id", claims.UserID, "session_id", claims.SessionID)

			response.Abort(c, http.StatusUnauthorized, constants.TokenInvalid, "Invalid access token")
			return
		}

		rlog.Debug(c, "authenticated request", "user_id", claims.UserID, "role", claims.Role)

		c.Set(constants.UserIDKey, claims.UserID)
		c.Set(constants.RoleKey, claims.Role)

		c.Next()
	}
}

// tokenRevoked checks for a revocation such as a logout or a password reset. If
// Redis fails it logs the error and allows the request: an outage shouldn't log out
// every user, and tokens still expire on their own within minutes.
func tokenRevoked(c *gin.Context, revoker session.TokenRevoker, claims *session.AccessClaims) bool {
	ctx, cancel := context.WithTimeout(c.Request.Context(), revocationCheckTimeout)
	defer cancel()

	revocations, err := revoker.Revocations(ctx, claims.UserID, claims.SessionID)
	if err != nil {
		rlog.Error(c, "token revocation check failed, allowing request", err, "user_id", claims.UserID)
		return false
	}
	return revocations.Rejects(claims)
}
