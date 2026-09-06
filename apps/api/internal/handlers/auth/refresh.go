package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Refresh(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenString, err := c.Cookie("refresh_token")
		if err != nil {
			rlog.Warn(c, "missing refresh token cookie")
			utils.ClearAuthCookies(c, cfg)
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Missing refresh token",
				Code:    constants.TokenNotProvided,
			})
			return
		}

		claims, err := session.ParseRefresh(refreshTokenString, cfg.JWTRefreshSecret)
		if err != nil {
			rlog.Warn(c, "invalid refresh token", "error", err)
			utils.ClearAuthCookies(c, cfg)
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid refresh token",
				Code:    constants.TokenInvalid,
			})
			return
		}

		familyID, err := utils.ConvertToUUID(claims.FamilyID)
		if err != nil {
			rlog.Warn(c, "invalid family_id in refresh token", "error", err)
			utils.ClearAuthCookies(c, cfg)
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		sess, err := queries.GetActiveSessionByID(ctx, familyID)
		if err != nil {
			rlog.Warn(c, "session not found or expired", "family_id", claims.FamilyID, "error", err)
			utils.ClearAuthCookies(c, cfg)
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Session expired",
				Code:    constants.TokenInvalid,
			})
			return
		}

		presentedHash := session.HashRefreshToken(refreshTokenString)

		user, err := queries.GetUserByID(ctx, sess.UserID)
		if err != nil {
			rlog.Error(c, "failed to fetch user", err, "user_id", sess.UserID)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		issueGraceAccess := func() {
			accessToken, err := session.NewAccessToken(cfg, user)
			if err != nil {
				rlog.Error(c, "failed to sign access token", err, "user_id", user.ID)
				c.JSON(http.StatusInternalServerError, types.APIResponse{
					Success: false,
					Message: "Failed to process request",
					Code:    constants.InternalServerError,
				})
				return
			}

			session.SetAccessCookie(c, cfg, accessToken)
			rlog.Info(c, "grace-window refresh (previous token reuse)", "user_id", user.ID)
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Tokens refreshed",
			})
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
				c.JSON(http.StatusInternalServerError, types.APIResponse{
					Success: false,
					Message: "Failed to process request",
					Code:    constants.InternalServerError,
				})
				return
			}

			_, err = queries.RotateSessionToken(ctx, db.RotateSessionTokenParams{
				ID:                sess.ID,
				NewTokenHash:      session.HashRefreshToken(tokens.RefreshToken),
				ExpectedTokenHash: presentedHash,
				ExpiresAt:         pgtype.Timestamptz{Time: time.Now().Add(cfg.RefreshTokenTTL), Valid: true},
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					// Lost the CAS race — another request already rotated this token.
					// Reload and serve the grace path instead of failing the client.
					reloaded, reloadErr := queries.GetActiveSessionByID(ctx, familyID)
					if reloadErr != nil {
						rlog.Warn(c, "session gone after lost rotation race", "error", reloadErr)
						utils.ClearAuthCookies(c, cfg)
						c.JSON(http.StatusUnauthorized, types.APIResponse{
							Success: false,
							Message: "Session expired",
							Code:    constants.TokenInvalid,
						})
						return
					}
					if withinGrace(reloaded) {
						issueGraceAccess()
						return
					}
					rlog.Warn(c, "lost rotation race but grace window missed", "user_id", user.ID)
					utils.ClearAuthCookies(c, cfg)
					c.JSON(http.StatusUnauthorized, types.APIResponse{
						Success: false,
						Message: "Session invalid, please log in again",
						Code:    constants.TokenInvalid,
					})
					return
				}

				rlog.Error(c, "failed to rotate session token", err, "user_id", user.ID)
				c.JSON(http.StatusInternalServerError, types.APIResponse{
					Success: false,
					Message: "Failed to process request",
					Code:    constants.InternalServerError,
				})
				return
			}

			session.SetSessionCookies(c, cfg, tokens.AccessToken, tokens.RefreshToken)
			rlog.Info(c, "tokens rotated successfully", "user_id", user.ID)
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Tokens refreshed",
			})

		case withinGrace(sess):
			issueGraceAccess()

		default:
			rlog.Warn(c, "stale or reused token outside grace window", "user_id", user.ID)
			if err := queries.RevokeSession(ctx, sess.ID); err != nil {
				rlog.Error(c, "failed to revoke session after stale token", err)
			}
			utils.ClearAuthCookies(c, cfg)
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Session invalid, please log in again",
				Code:    constants.TokenInvalid,
			})
		}
	}
}
