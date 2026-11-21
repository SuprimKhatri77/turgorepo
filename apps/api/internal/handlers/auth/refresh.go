package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Refresh(queries Store, cfg *config.Config, revoker session.TokenRevoker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenString, err := c.Cookie(session.RefreshTokenCookie)
		if err != nil {
			rlog.Warn(c, "missing refresh token cookie")
			session.ClearCookies(c, cfg)
			response.Error(c, http.StatusUnauthorized, constants.TokenNotProvided, "Missing refresh token")
			return
		}

		claims, err := session.ParseRefresh(refreshTokenString, cfg.JWTRefreshSecret)
		if err != nil {
			rlog.Warn(c, "invalid refresh token", "error", err)
			session.ClearCookies(c, cfg)
			response.Error(c, http.StatusUnauthorized, constants.TokenInvalid, "Invalid refresh token")
			return
		}

		familyID, err := utils.ConvertToUUID(claims.FamilyID)
		if err != nil {
			rlog.Warn(c, "invalid family_id in refresh token", "error", err)
			session.ClearCookies(c, cfg)
			response.Error(c, http.StatusUnauthorized, constants.InvalidToken, "Invalid token claims")
			return
		}

		sess, err := queries.GetActiveSessionByID(ctx, familyID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				respondSessionLookupFailed(c, err)
				return
			}
			rlog.Warn(c, "session not found or expired", "family_id", claims.FamilyID)
			session.ClearCookies(c, cfg)
			response.Error(c, http.StatusUnauthorized, constants.TokenInvalid, "Session expired")
			return
		}

		presentedHash := session.HashToken(refreshTokenString)

		user, err := queries.GetUserByID(ctx, sess.UserID)
		if err != nil {
			rlog.Error(c, "failed to fetch user", err, "user_id", sess.UserID)
			response.Internal(c)
			return
		}

		issueGraceAccess := func() {
			accessToken, err := session.NewAccessToken(cfg, user, familyID)
			if err != nil {
				rlog.Error(c, "failed to sign access token", err, "user_id", user.ID)
				response.Internal(c)
				return
			}

			session.SetAccessCookie(c, cfg, accessToken)
			rlog.Info(c, "grace-window refresh (previous token reuse)", "user_id", user.ID)
			response.OK(c, http.StatusOK, "Tokens refreshed", nil)
		}

		withinGrace := func(s db.Session) bool {
			return s.PreviousTokenHash.Valid &&
				presentedHash == s.PreviousTokenHash.String &&
				s.PreviousRotatedAt.Valid &&
				time.Since(s.PreviousRotatedAt.Time) < cfg.RefreshReuseWindow
		}

		switch {
		case presentedHash == sess.CurrentTokenHash:
			tokens, err := session.NewTokens(cfg, user, familyID)
			if err != nil {
				rlog.Error(c, "failed to sign tokens", err, "user_id", user.ID)
				response.Internal(c)
				return
			}

			_, err = queries.RotateSessionToken(ctx, db.RotateSessionTokenParams{
				ID:                sess.ID,
				NewTokenHash:      session.HashToken(tokens.RefreshToken),
				ExpectedTokenHash: presentedHash,
				ExpiresAt:         pgtype.Timestamptz{Time: time.Now().Add(cfg.RefreshTokenTTL), Valid: true},
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					// Lost the CAS race — another request already rotated this token.
					// Reload and serve the grace path instead of failing the client.
					reloaded, reloadErr := queries.GetActiveSessionByID(ctx, familyID)
					if reloadErr != nil {
						if !errors.Is(reloadErr, pgx.ErrNoRows) {
							respondSessionLookupFailed(c, reloadErr)
							return
						}
						rlog.Warn(c, "session gone after lost rotation race")
						session.ClearCookies(c, cfg)
						response.Error(c, http.StatusUnauthorized, constants.TokenInvalid, "Session expired")
						return
					}
					if withinGrace(reloaded) {
						issueGraceAccess()
						return
					}
					rlog.Warn(c, "lost rotation race but grace window missed", "user_id", user.ID)
					session.ClearCookies(c, cfg)
					response.Error(c, http.StatusUnauthorized, constants.TokenInvalid, "Session invalid, please log in again")
					return
				}

				rlog.Error(c, "failed to rotate session token", err, "user_id", user.ID)
				response.Internal(c)
				return
			}

			session.SetCookies(c, cfg, tokens.AccessToken, tokens.RefreshToken)
			rlog.Info(c, "tokens rotated successfully", "user_id", user.ID)
			response.OK(c, http.StatusOK, "Tokens refreshed", nil)

		case withinGrace(sess):
			issueGraceAccess()

		default:
			// A reused refresh token may have been stolen, so end the session, including
			// access tokens already issued for it, whoever holds them.
			rlog.Warn(c, "stale or reused token outside grace window", "user_id", user.ID)
			if err := queries.RevokeSession(ctx, sess.ID); err != nil {
				rlog.Error(c, "failed to revoke session after stale token", err)
			}
			revokeSessionTokens(c, revoker, cfg, claims.FamilyID)
			session.ClearCookies(c, cfg)
			response.Error(c, http.StatusUnauthorized, constants.TokenInvalid, "Session invalid, please log in again")
		}
	}
}

// respondSessionLookupFailed reports a database failure without clearing cookies:
// the session may still be valid, so an outage must not log the user out.
func respondSessionLookupFailed(c *gin.Context, err error) {
	rlog.Error(c, "failed to look up session", err)
	response.Error(c, http.StatusServiceUnavailable, constants.ServiceUnavailable, "Service temporarily unavailable")
}
