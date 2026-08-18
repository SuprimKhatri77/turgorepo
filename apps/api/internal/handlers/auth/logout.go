package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
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

		userID, err := utils.ConvertToUUID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusOK, types.APIResponse{
				Success: true,
				Message: "Logged out successfully",
			})
			return
		}

		_, err = queries.RevokeTokenByUserIDAndToken(ctx, db.RevokeTokenByUserIDAndTokenParams{
			Token:  session.HashRefreshToken(refreshTokenFromCookie),
			UserID: userID,
		})
		if err != nil {
			rlog.Error(c, "failed to revoke refresh token on logout", err)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to logout",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "user logged out", "user_id", claims.UserID)

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out successfully",
		})
	}
}
