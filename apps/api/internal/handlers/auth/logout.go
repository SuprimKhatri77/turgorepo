package auth

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

		refreshTokenFromCookie, err := c.Cookie("refresh_token")
		if err != nil {
			rlog.Warn(c, "missing refresh token on logout")

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Missing refresh token",
				Code:    constants.TokenNotProvided,
			})
			return
		}

		token, err := jwt.Parse(refreshTokenFromCookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				rlog.Error(c, "unexpected signing method during logout", fmt.Errorf("unexpected signing method: %v", token.Header["alg"]), "alg", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTRefreshSecret), nil
		})

		if err != nil || !token.Valid {
			rlog.Warn(c, "invalid refresh token on logout", "error", err)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid refresh token",
				Code:    constants.TokenInvalid,
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			rlog.Warn(c, "invalid token claims on logout")

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		userID, err := utils.ConvertToUUID(claims["user_id"].(string))
		if err != nil {
			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid token claims",
				Code:    constants.InvalidToken,
			})
			return
		}

		hash := sha256.Sum256([]byte(refreshTokenFromCookie))
		tokenHash := fmt.Sprintf("%x", hash)

		result, err := queries.RevokeTokenByUserIDAndToken(ctx, db.RevokeTokenByUserIDAndTokenParams{
			Token:  tokenHash,
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
		if result.RowsAffected() == 0 {
			rlog.Warn(c, "refresh token not found on logout")

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Refresh token not found",
				Code:    constants.TokenInvalid,
			})
			return
		}

		utils.SetAuthCookie(c, "access_token", "", -1, cfg)
		utils.SetAuthCookie(c, "refresh_token", "", -1, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "", -1, cfg)

		rlog.Info(c, "user logged out")

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Logged out successfully",
		})
	}
}
