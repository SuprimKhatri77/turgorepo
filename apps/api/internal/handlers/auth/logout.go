package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Logout(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		refreshTokenFromCookie, cookieErr := c.Cookie("refresh_token")
		utils.ClearAuthCookies(c, cfg)

		if cookieErr != nil || refreshTokenFromCookie == "" {
			rlog.Info(c, "logout without refresh token")
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Logged out successfully",
			})
			return
		}

		claims, err := session.ParseRefresh(refreshTokenFromCookie, cfg.JWTRefreshSecret)
		if err != nil {
			rlog.Info(c, "logout with invalid refresh token")
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Logged out successfully",
			})
			return
		}

		familyID, err := utils.ConvertToUUID(claims.FamilyID)
		if err != nil {
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Logged out successfully",
			})
			return
		}

		// Best-effort revoke: cookies are already cleared, so the client is logged out.
		// Log and still return 200 — failing logout after clearing cookies would confuse the user.
		if err := queries.RevokeSession(ctx, familyID); err != nil {
			rlog.Error(c, "failed to revoke session on logout", err)
		}

		rlog.Info(c, "user logged out", "user_id", claims.UserID)
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out successfully",
		})
	}
}
