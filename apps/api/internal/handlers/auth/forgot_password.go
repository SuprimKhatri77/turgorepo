package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

type ForgotPasswordRequest struct {
	Email string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
}

// ForgotPassword emails a reset link. Everything after validation happens in the
// background, so the response and its timing are the same whether or not the
// account exists, and whether or not sending works.
func (h *Handler) ForgotPassword(c *gin.Context) {
	req, ok := validator.BindJSON[ForgotPasswordRequest](c)
	if !ok {
		return
	}

	emailAddress := utils.NormalizeEmail(req.Email)
	h.sendInBackground(c, "password reset email", func(ctx context.Context) error {
		return h.sendPasswordResetEmail(ctx, emailAddress)
	})

	response.OK(c, http.StatusOK, "If that email has an account, we've sent a reset link", nil)
}

func (h *Handler) sendPasswordResetEmail(ctx context.Context, emailAddress string) error {
	// Limited per address across all IPs, so nobody can flood someone's inbox.
	if !ratelimit.Allowed(ctx, h.Limiter, ratelimit.ForgotPasswordPerEmail, ratelimit.HashEmail(emailAddress)) {
		return nil
	}

	user, err := h.Store.GetUserByEmail(ctx, emailAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	token, tokenHash, err := session.NewEmailToken()
	if err != nil {
		return err
	}

	ttl := h.Config.PasswordResetTTL
	if err := h.Store.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true},
	}); err != nil {
		return err
	}

	message, err := email.PasswordResetEmail(user.Email, user.Name, frontendLink(h.Config, "/auth/reset-password", token), ttl)
	if err != nil {
		return err
	}
	return h.Mailer.Send(ctx, message)
}
