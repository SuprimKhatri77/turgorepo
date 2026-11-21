package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordRequest struct {
	Token    string `json:"token" label:"Reset link" binding:"required,max=128" msg_required:"Reset link is missing" msg_max:"Reset link is invalid"`
	Password string `json:"password" trim:"-" binding:"required,not_blank,min=8,max_bytes=72"`
}

// ResetPassword sets a new password using a token from a reset email, and logs the
// user out everywhere, including access tokens that haven't expired yet.
func (h *Handler) ResetPassword(c *gin.Context) {
	req, ok := validator.BindJSON[ResetPasswordRequest](c)
	if !ok {
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		rlog.Error(c, "failed to hash password", err)
		response.Internal(c)
		return
	}

	user, err := h.Store.ResetPasswordWithToken(c.Request.Context(), db.ResetPasswordWithTokenParams{
		TokenHash:    session.HashToken(req.Token),
		PasswordHash: string(passwordHash),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		rlog.Warn(c, "invalid or expired password reset token")
		response.Error(c, http.StatusBadRequest, constants.InvalidResetToken, "This reset link is invalid or has expired")
		return
	}
	if err != nil {
		rlog.Error(c, "failed to reset password", err)
		response.Internal(c)
		return
	}

	rlog.Info(c, "password reset", "user_id", user.ID)

	h.revokeAccessTokens(c, user)

	// Someone locked out by wrong passwords should be able to log in straight away.
	ratelimit.ClearAccount(c, h.Limiter, user.Email)
	session.ClearCookies(c, h.Config)

	h.sendInBackground(c, "password changed email", func(ctx context.Context) error {
		return h.sendPasswordChangedEmail(ctx, user)
	})

	response.OK(c, http.StatusOK, "Your password has been reset. Please sign in", nil)
}

// revokeAccessTokens runs after the database commit. A failure doesn't fail the
// reset, because the password and sessions have already changed.
func (h *Handler) revokeAccessTokens(c *gin.Context, user db.User) {
	userID, err := utils.UUIDString(user.ID)
	if err != nil {
		rlog.Error(c, "failed to revoke access tokens after password reset", err, "user_id", user.ID)
		return
	}
	h.revokeUserTokens(c, userID)
}

func (h *Handler) sendPasswordChangedEmail(ctx context.Context, user db.User) error {
	message, err := email.PasswordChangedEmail(user.Email, user.Name)
	if err != nil {
		return err
	}
	return h.Mailer.Send(ctx, message)
}
