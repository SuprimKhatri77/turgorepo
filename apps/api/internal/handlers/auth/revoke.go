package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

// revokeTimeout bounds the Redis write that revokes access tokens, so a slow
// Redis can't hold up logout or a password reset.
const revokeTimeout = 500 * time.Millisecond

// sessionFromCookies finds the session to end. The refresh token names it, and a
// valid access token is the fallback, so a client that lost its refresh cookie can
// still revoke the access token it holds.
func sessionFromCookies(c *gin.Context, cfg *config.Config) (sessionID, userID string, found bool) {
	if token, err := c.Cookie("refresh_token"); err == nil && token != "" {
		if claims, err := session.ParseRefresh(token, cfg.JWTRefreshSecret); err == nil {
			return claims.FamilyID, claims.UserID, true
		}
	}
	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		if claims, err := session.ParseAccess(token, cfg.JWTAccessSecret); err == nil {
			return claims.SessionID, claims.UserID, true
		}
	}
	return "", "", false
}

// revokeSessionTokens rejects the session's access tokens until they expire. A
// failure is logged rather than returned: the session is already revoked in the
// database, so its tokens can't be refreshed and expire within minutes.
func revokeSessionTokens(c *gin.Context, revoker session.TokenRevoker, cfg *config.Config, sessionID string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), revokeTimeout)
	defer cancel()

	if err := revoker.RevokeSession(ctx, sessionID, cfg.AccessTokenTTL); err != nil {
		rlog.Error(c, "failed to revoke session access tokens", err, "session_id", sessionID)
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
