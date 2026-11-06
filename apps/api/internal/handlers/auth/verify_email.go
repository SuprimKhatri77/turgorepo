package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
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
func VerifyEmail(deps AccountEmailDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, ok := validator.BindJSON[VerifyEmailRequest](c)
		if !ok {
			return
		}

		user, err := deps.Queries.VerifyEmailWithToken(c.Request.Context(), session.HashToken(req.Token))
		if errors.Is(err, pgx.ErrNoRows) {
			rlog.Warn(c, "invalid or expired email verification token")
			c.JSON(http.StatusBadRequest, types.APIResponse{
				Success: false,
				Message: "This verification link is invalid or has expired",
				Code:    constants.InvalidVerificationToken,
			})
			return
		}
		if err != nil {
			rlog.Error(c, "failed to verify email", err)
			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to process request",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "email verified", "user_id", user.ID)
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Your email is verified. Please sign in",
		})
	}
}

// ResendVerification emails a new verification link. Like ForgotPassword, the
// work happens in the background, so the response is the same whether or not
// the account exists or is already verified.
func ResendVerification(deps AccountEmailDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, ok := validator.BindJSON[ResendVerificationRequest](c)
		if !ok {
			return
		}

		emailAddress := utils.NormalizeEmail(req.Email)
		requestID := c.GetString(constants.RequestIDKey)

		deps.Jobs.Go("verification email", func(ctx context.Context) {
			if err := resendVerificationEmail(ctx, deps, emailAddress); err != nil {
				slog.Error("failed to send verification email", "request_id", requestID, "err", err)
			}
		})

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "If that email has an unverified account, we've sent a new link",
		})
	}
}

func resendVerificationEmail(ctx context.Context, deps AccountEmailDeps, emailAddress string) error {
	// Limited per address across all IPs, so nobody can flood someone's inbox.
	if !ratelimit.Allowed(ctx, deps.Limiter, ratelimit.ResendVerificationPerEmail, ratelimit.HashEmail(emailAddress)) {
		return nil
	}

	user, err := deps.Queries.GetUserByEmail(ctx, emailAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt.Valid {
		return nil
	}
	return sendVerificationEmail(ctx, deps, user)
}

// sendVerificationEmail stores a new token, which makes older links stop
// working, and emails it.
func sendVerificationEmail(ctx context.Context, deps AccountEmailDeps, user db.User) error {
	token, tokenHash, err := session.NewEmailToken()
	if err != nil {
		return err
	}

	ttl := deps.Config.EmailVerificationTTL
	if err := deps.Queries.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true},
	}); err != nil {
		return err
	}

	message, err := email.VerificationEmail(user.Email, user.Name, verifyURL(deps.Config, token), ttl)
	if err != nil {
		return err
	}
	return deps.Mailer.Send(ctx, message)
}

// verifyURL builds the link the same way as resetURL: from configuration, with
// the token in the fragment.
func verifyURL(cfg *config.Config, token string) string {
	return strings.TrimRight(cfg.FrontendURL, "/") + "/auth/verify-email#token=" + token
}
