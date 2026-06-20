package auth

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Logout(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenFromCookie, err := c.Cookie("refresh_token")
		if err != nil {
			slog.Warn("missing refresh token on logout",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Missing refresh token",
				Code:    constants.TokenNotProvided,
			})
			return
		}

		token, err := jwt.Parse(refreshTokenFromCookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				slog.Error("unexpected signing method during logout",
					"alg", token.Header["alg"],
				)
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTRefreshSecret), nil
		})

		if err != nil || !token.Valid {
			slog.Warn("invalid refresh token on logout",
				"error", err,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid refresh token",
				Code:    constants.TokenInvalid,
			})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			slog.Warn("invalid token claims",
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token",
				Code:    constants.InvalidToken,
			})
			return
		}

		sessionIDFromClaims, ok := claims["session_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		sessionID, err := utils.ConvertToUUID(sessionIDFromClaims)
		if err != nil {
			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		hash := sha256.Sum256([]byte(refreshTokenFromCookie))
		tokenHash := fmt.Sprintf("%x", hash)

		_, err = queries.RevokeTokenBySessionIDAndToken(ctx, db.RevokeTokenBySessionIDAndTokenParams{
			Token:     tokenHash,
			SessionID: sessionID,
		})
		if err != nil {
			slog.Error("failed to revoke refresh token on logout",
				"error", err,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to logout",
				Code:    constants.InternalServerError,
			})
			return
		}

		utils.SetAuthCookie(c, "access_token", "", -1, cfg)
		utils.SetAuthCookie(c, "refresh_token", "", -1, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "", -1, cfg)

		slog.Info("user logged out",
			"path", c.FullPath(),
			"ip", c.ClientIP(),
		)

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out successfully",
		})
	}
}
