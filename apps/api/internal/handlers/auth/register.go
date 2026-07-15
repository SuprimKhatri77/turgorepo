package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,not_blank,min=2,max=50,alphaspace"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,not_blank,min=8,max=50"`
}

func Register(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		jti := uuid.New()

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			rlog.Warn(c, "invalid request payload", "error", err)

			c.JSON(http.StatusBadRequest, types.APIResponse{
				Success: false,
				Message: "Invalid request body",
				Code:    constants.ValidationFailed,
				Errors:  validator.Parse(err, req),
			})
			return
		}

		utils.TrimStruct(&req, "Password")

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			rlog.Error(c, "failed to hash password", err)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		user, err := queries.CreateUser(ctx, db.CreateUserParams{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(passwordHash),
			Role:         "member",
		})

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				rlog.Warn(c, "user already exists", "email", req.Email)

				c.JSON(http.StatusConflict, types.APIResponse{
					Success: false,
					Message: "User already exists",
					Code:    constants.UserAlreadyExists,
				})
				return
			}

			rlog.Error(c, "failed to create user", err)

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
			"jti":       jti,
			"exp":       time.Now().Add(15 * time.Minute).Unix(),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
		accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))
		if err != nil {
			rlog.Error(c, "failed to sign access token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		refreshClaims := jwt.MapClaims{
			"user_id": user.ID,
			"jti":     jti,
			"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
		if err != nil {
			rlog.Error(c, "failed to sign refresh token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		expiresAt := pgtype.Timestamptz{
			Time:  time.Now().Add(30 * 24 * time.Hour),
			Valid: true,
		}

		refreshTokenHash := sha256.Sum256([]byte(refreshTokenString))
		refreshTokenHashString := fmt.Sprintf("%x", refreshTokenHash)

		_, err = queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			UserID:    user.ID,
			ExpiresAt: expiresAt,
			Token:     refreshTokenHashString,
		})
		if err != nil {
			rlog.Error(c, "failed to store refresh token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "tokens issued", "user_id", user.ID)

		utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
		utils.SetAuthCookie(c, "refresh_token", refreshTokenString, 30*24*60*60, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "true", 30*24*60*60, cfg)

		rlog.Info(c, "registration successful", "user_id", user.ID)

		c.JSON(http.StatusCreated, types.APIResponse{
			Success: true,
			Message: "Registration successful",
			Data: db.User{
				ID:       user.ID,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
				ImageUrl: user.ImageUrl,
			},
		})

	}
}
