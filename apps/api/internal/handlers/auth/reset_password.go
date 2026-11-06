package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

const revokeTimeout = 500 * time.Millisecond

type ResetPasswordRequest struct {
	Token    string `json:"token" label:"Reset link" binding:"required,max=128" msg_required:"Reset link is missing" msg_max:"Reset link is invalid"`
	Password string `json:"password" label:"Password" binding:"required,not_blank,min=8,max_bytes=72" msg_required:"Password is required" msg_min:"Password must be at least 8 characters" msg_max_bytes:"Password cannot exceed 72 bytes" msg_not_blank:"Password cannot be blank"`
}

// ResetPassword sets a new password using a token from a reset email, and logs the
// user out everywhere, including access tokens that haven't expired yet.
func ResetPassword(deps AccountEmailDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, ok := validator.BindJSON[ResetPasswordRequest](c)
		if !ok {
			return
		}

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

		user, err := deps.Queries.ResetPasswordWithToken(c.Request.Context(), db.ResetPasswordWithTokenParams{
			TokenHash:    session.HashToken(req.Token),
			PasswordHash: string(passwordHash),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			rlog.Warn(c, "invalid or expired password reset token")
			c.JSON(http.StatusBadRequest, types.APIResponse{
				Success: false,
				Message: "This reset link is invalid or has expired",
				Code:    constants.InvalidResetToken,
			})
			return
		}
		if err != nil {
			rlog.Error(c, "failed to reset password", err)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "password reset", "user_id", user.ID)

		revokeAccessTokens(c, deps, user)

		// Someone locked out by wrong passwords should be able to log in straight away.
		ratelimit.ClearAccount(c, deps.Limiter, user.Email)
		utils.ClearAuthCookies(c, deps.Config)

		requestID := c.GetString(constants.RequestIDKey)
		deps.Jobs.Go("password changed email", func(ctx context.Context) {
			if err := sendPasswordChangedEmail(ctx, deps, user); err != nil {
				slog.Error("failed to send password changed email", "request_id", requestID, "err", err)
			}
		})

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Your password has been reset. Please sign in",
		})
	}
}

// revokeAccessTokens runs after the database commit. A failure doesn't fail the
// reset, because the password and sessions have already changed.
func revokeAccessTokens(c *gin.Context, deps AccountEmailDeps, user db.User) {
	userID, err := utils.UUIDString(user.ID)
	if err != nil {
		rlog.Error(c, "failed to revoke access tokens after password reset", err, "user_id", user.ID)
		return
	}
	revokeUserTokens(c, deps.Revoker, deps.Config, userID)
}

func sendPasswordChangedEmail(ctx context.Context, deps AccountEmailDeps, user db.User) error {
	message, err := email.PasswordChangedEmail(user.Email, user.Name)
	if err != nil {
		return err
	}
	return deps.Mailer.Send(ctx, message)
}
