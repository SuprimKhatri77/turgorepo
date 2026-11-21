package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

type VerifyEmailRequest struct {
	Token string `json:"token" label:"Verification link" binding:"required,max=128" msg_required:"Verification link is missing" msg_max:"Verification link is invalid"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
}

// VerifyEmail confirms an address using the token from a verification email. It
// doesn't log the user in, so the link never works as a password.
func (h *Handler) VerifyEmail(c *gin.Context) {
	req, ok := validator.BindJSON[VerifyEmailRequest](c)
	if !ok {
		return
	}

	user, err := h.Store.VerifyEmailWithToken(c.Request.Context(), session.HashToken(req.Token))
	if errors.Is(err, pgx.ErrNoRows) {
		rlog.Warn(c, "invalid or expired email verification token")
		response.Error(c, http.StatusBadRequest, constants.InvalidVerificationToken, "This verification link is invalid or has expired")
		return
	}
	if err != nil {
		rlog.Error(c, "failed to verify email", err)
		response.Internal(c)
		return
	}

	rlog.Info(c, "email verified", "user_id", user.ID)
	response.OK(c, http.StatusOK, "Your email is verified. Please sign in", nil)
}

// ResendVerification emails a new verification link. Like ForgotPassword, the
// work happens in the background, so the response is the same whether or not
// the account exists or is already verified.
func (h *Handler) ResendVerification(c *gin.Context) {
	req, ok := validator.BindJSON[ResendVerificationRequest](c)
	if !ok {
		return
	}

	emailAddress := utils.NormalizeEmail(req.Email)
	h.sendInBackground(c, "verification email", func(ctx context.Context) error {
		return h.resendVerificationEmail(ctx, emailAddress)
	})

	response.OK(c, http.StatusOK, "If that email has an unverified account, we've sent a new link", nil)
}

func (h *Handler) resendVerificationEmail(ctx context.Context, emailAddress string) error {
	// Limited per address across all IPs, so nobody can flood someone's inbox.
	if !ratelimit.Allowed(ctx, h.Limiter, ratelimit.ResendVerificationPerEmail, ratelimit.HashEmail(emailAddress)) {
		return nil
	}

	user, err := h.Store.GetUserByEmail(ctx, emailAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt.Valid {
		return nil
	}
	return h.sendVerificationEmail(ctx, user)
}

// sendVerificationEmail stores a new token, which makes older links stop
// working, and emails it.
func (h *Handler) sendVerificationEmail(ctx context.Context, user db.User) error {
	token, tokenHash, err := session.NewEmailToken()
	if err != nil {
		return err
	}

	ttl := h.Config.EmailVerificationTTL
	if err := h.Store.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true},
	}); err != nil {
		return err
	}

	message, err := email.VerificationEmail(user.Email, user.Name, frontendLink(h.Config, "/auth/verify-email", token), ttl)
	if err != nil {
		return err
	}
	return h.Mailer.Send(ctx, message)
}
