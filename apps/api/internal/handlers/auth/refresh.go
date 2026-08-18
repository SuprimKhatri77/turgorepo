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

		userID, err := utils.ConvertToUUID(claims.UserID)
		if err != nil {
			rlog.Warn(c, "invalid user_id in refresh token", "error", err)
			utils.ClearAuthCookies(c, cfg)
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		refreshTokenHash := session.HashRefreshToken(refreshTokenString)
		stored, err := queries.GetRefreshTokenByUserIDAndToken(ctx, db.GetRefreshTokenByUserIDAndTokenParams{
			UserID: userID,
			Token:  refreshTokenHash,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				utils.ClearAuthCookies(c, cfg)
				c.JSON(http.StatusUnauthorized, types.APIResponse{
					Success: false,
					Message: "Invalid refresh token",
					Code:    constants.TokenInvalid,
				})
				return
			}

			rlog.Error(c, "failed to fetch refresh token", err)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		user, err := queries.GetUserByID(ctx, stored.UserID)
		if err != nil {
			rlog.Error(c, "failed to fetch user", err, "user_id", stored.UserID)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		tokens, err := session.NewTokens(cfg, user)
		if err != nil {
			rlog.Error(c, "failed to sign tokens", err, "user_id", user.ID)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		if time.Since(stored.CreatedAt.Time) < cfg.RefreshReuseWindow {
			session.SetAccessCookie(c, cfg, tokens.AccessToken)
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Tokens refreshed",
			})
			return
		}

		_, err = queries.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
			UserID:   userID,
			OldToken: refreshTokenHash,
			NewToken: session.HashRefreshToken(tokens.RefreshToken),
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(cfg.RefreshTokenTTL),
				Valid: true,
			},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				rlog.Warn(c, "refresh token already rotated", "user_id", user.ID)
				utils.ClearAuthCookies(c, cfg)
				c.JSON(http.StatusUnauthorized, types.APIResponse{
					Success: false,
					Message: "Invalid refresh token",
					Code:    constants.TokenInvalid,
				})
				return
			}

			rlog.Error(c, "failed to rotate refresh token", err, "user_id", user.ID)
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
	}
}
