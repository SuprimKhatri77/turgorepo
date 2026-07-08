package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			slog.Warn("missing access token",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Missing access token",
				Code:    constants.TokenNotProvided,
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				slog.Error("unexpected signing method",
					"alg", token.Header["alg"],
				)
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTAccessSecret), nil
		})

		if err != nil || !token.Valid {
			slog.Warn("invalid access token",
				"error", err,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid access token",
				Code:    constants.TokenInvalid,
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			slog.Warn("invalid token claims structure",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.Unauthorized,
			})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			slog.Warn("missing userID in claims",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.Unauthorized,
			})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok || (role != "admin" && role != "superadmin") {
			slog.Warn("invalid role in claims",
				"user_id", userID,
				"role", role,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.Unauthorized,
			})
			c.Abort()
			return
		}

		slog.Info("authenticated request",
			"user_id", userID,
			"role", role,
			"path", c.FullPath(),
			"ip", c.ClientIP(),
		)

		c.Set("user_id", userID)
		c.Set("role", role)

		c.Next()
	}
}
