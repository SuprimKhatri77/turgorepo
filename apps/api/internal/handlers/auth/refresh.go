package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Refresh(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenString, err := c.Cookie("refresh_token")
		if err != nil {
			handlerlog.Warn(c, "missing refresh token cookie")

			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusBadRequest, types.APIResponse{
				Success: false,
				Message: "Missing refresh token",
				Code:    constants.TokenNotProvided,
			})
			return
		}

		token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				handlerlog.Error(c, "unexpected signing method", fmt.Errorf("unexpected signing method: %v", token.Header["alg"]), "alg", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTRefreshSecret), nil
		})

		if err != nil || !token.Valid {
			handlerlog.Warn(c, "invalid refresh token", "error", err)

			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid refresh token",
				Code:    constants.TokenInvalid,
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			handlerlog.Warn(c, "invalid token claims")

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token",
				Code:    constants.InvalidToken,
			})
			return
		}

		sessionIDFromClaims, ok := claims["session_id"].(string)
		if !ok {
			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		handlerlog.Info(c, "session from claims", "session_id", sessionIDFromClaims)

		sessionID, err := utils.ConvertToUUID(sessionIDFromClaims)
		if err != nil {
			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		refreshTokenHash := sha256.Sum256([]byte(refreshTokenString))
		refreshTokenHashString := fmt.Sprintf("%x", refreshTokenHash)

		refreshToken, err := queries.GetRefreshTokenBySessionIDAndToken(ctx, db.GetRefreshTokenBySessionIDAndTokenParams{
			SessionID: sessionID,
			Token:     refreshTokenHashString,
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

			handlerlog.Error(c, "failed to fetch refresh token", err)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		user, err := queries.GetUserByID(ctx, refreshToken.UserID)
		if err != nil {
			handlerlog.Error(c, "failed to fetch user", err, "user_id", refreshToken.UserID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		accessClaims := jwt.MapClaims{
			"user_id":   user.ID,
			"role":      user.Role,
			"email":     user.Email,
			"name":      user.Name,
			"image_url": user.ImageUrl,
			"exp":       time.Now().Add(15 * time.Minute).Unix(),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
		accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign access token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		if time.Since(refreshToken.CreatedAt.Time) < 5*time.Minute {
			utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
			})
			return
		}

		_, err = queries.RevokeTokenBySessionIDAndToken(ctx, db.RevokeTokenBySessionIDAndTokenParams{
			SessionID: sessionID,
			Token:     refreshTokenHashString,
		})
		if err != nil {
			handlerlog.Error(c, "failed to revoke refresh token", err, "user_id", refreshToken.UserID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		refreshClaims := jwt.MapClaims{
			"user_id":    user.ID,
			"session_id": sessionID,
			"exp":        time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		newRefreshTokenString, err := newRefreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign refresh token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		newHash := sha256.Sum256([]byte(newRefreshTokenString))
		newTokenHash := fmt.Sprintf("%x", newHash)

		_, err = queries.CreateToken(ctx, db.CreateTokenParams{
			UserID: user.ID,
			Token:  newTokenHash,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(30 * 24 * time.Hour),
				Valid: true,
			},
			SessionID: sessionID,
		})
		if err != nil {
			handlerlog.Error(c, "failed to persist new refresh token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
		utils.SetAuthCookie(c, "refresh_token", newRefreshTokenString, 30*24*60*60, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "true", 30*24*60*60, cfg)

		handlerlog.Info(c, "tokens rotated successfully", "user_id", user.ID)

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Tokens refreshed",
		})
	}
}
