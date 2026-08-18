package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			rlog.Warn(c, "missing access token")

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Missing access token",
				Code:    constants.TokenNotProvided,
			})
			c.Abort()
			return
		}

		claims, err := session.ParseAccess(accessToken, cfg.JWTAccessSecret)
		if err != nil {
			rlog.Warn(c, "invalid access token", "error", err)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid access token",
				Code:    constants.TokenInvalid,
			})
			c.Abort()
			return
		}

		if !constants.IsValidRole(claims.Role) {
			rlog.Warn(c, "invalid role in claims", "user_id", claims.UserID, "role", claims.Role)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.Unauthorized,
			})
			c.Abort()
			return
		}

		rlog.Info(c, "authenticated request", "user_id", claims.UserID, "role", claims.Role)

		c.Set(constants.UserIDKey, claims.UserID)
		c.Set(constants.RoleKey, claims.Role)

		c.Next()
	}
}
