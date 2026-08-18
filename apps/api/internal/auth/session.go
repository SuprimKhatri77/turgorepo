package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func IssueSession(c *gin.Context, queries repository.AuthRepository, cfg *config.Config, user db.User) error {
	tokens, err := NewTokens(cfg, user)
	if err != nil {
		return err
	}

	_, err = queries.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
		UserID: user.ID,
		Token:  HashRefreshToken(tokens.RefreshToken),
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(cfg.RefreshTokenTTL),
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	SetSessionCookies(c, cfg, tokens.AccessToken, tokens.RefreshToken)
	return nil
}

func SetSessionCookies(c *gin.Context, cfg *config.Config, accessToken, refreshToken string) {
	utils.SetAuthCookie(c, "access_token", accessToken, cfg.AccessCookieMaxAge(), cfg)
	utils.SetAuthCookie(c, "refresh_token", refreshToken, cfg.RefreshCookieMaxAge(), cfg)
	utils.SetPublicCookie(c, "is_logged_in", "true", cfg.RefreshCookieMaxAge(), cfg)
}

func SetAccessCookie(c *gin.Context, cfg *config.Config, accessToken string) {
	utils.SetAuthCookie(c, "access_token", accessToken, cfg.AccessCookieMaxAge(), cfg)
}

func PublicUser(user db.User) db.User {
	return db.User{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		ImageUrl: user.ImageUrl,
	}
}
