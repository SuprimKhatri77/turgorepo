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
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

type ForgotPasswordRequest struct {
	Email string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
}

// ForgotPassword emails a reset link. Everything after validation happens in the
// background, so the response and its timing are the same whether or not the
// account exists, and whether or not sending works.
func ForgotPassword(deps AccountEmailDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, ok := validator.BindJSON[ForgotPasswordRequest](c)
		if !ok {
			return
		}

		emailAddress := utils.NormalizeEmail(req.Email)
		requestID := c.GetString(constants.RequestIDKey)

		deps.Jobs.Go("password reset email", func(ctx context.Context) {
			if err := sendPasswordResetEmail(ctx, deps, emailAddress); err != nil {
				slog.Error("failed to send password reset email", "request_id", requestID, "err", err)
			}
		})

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "If that email has an account, we've sent a reset link",
		})
	}
}

func sendPasswordResetEmail(ctx context.Context, deps AccountEmailDeps, emailAddress string) error {
	// Limited per address across all IPs, so nobody can flood someone's inbox.
	if !ratelimit.Allowed(ctx, deps.Limiter, ratelimit.ForgotPasswordPerEmail, ratelimit.HashEmail(emailAddress)) {
		return nil
	}

	user, err := deps.Queries.GetUserByEmail(ctx, emailAddress)
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

	ttl := deps.Config.PasswordResetTTL
	if err := deps.Queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true},
	}); err != nil {
		return err
	}

	message, err := email.PasswordResetEmail(user.Email, user.Name, resetURL(deps.Config, token), ttl)
	if err != nil {
		return err
	}
	return deps.Mailer.Send(ctx, message)
}

// resetURL uses configuration rather than request headers, so a forged Host can't
// send the token to someone else's site. The token is in the fragment because
// browsers don't send fragments to servers, which keeps it out of access logs.
func resetURL(cfg *config.Config, token string) string {
	return strings.TrimRight(cfg.FrontendURL, "/") + "/auth/reset-password#token=" + token
}
