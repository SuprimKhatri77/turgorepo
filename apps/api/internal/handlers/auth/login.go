package auth

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
	Password string `json:"password" trim:"-" label:"Password" binding:"required,max_bytes=72" msg_required:"Password is required" msg_max_bytes:"Password cannot exceed 72 bytes"`
}

// dummyPasswordHash is compared against when the email is unknown, so that path
// takes as long as a wrong password and response time doesn't reveal which emails exist.
var dummyPasswordHash = sync.OnceValue(func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
})

func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req, ok := validator.BindJSON[LoginRequest](c)
	if !ok {
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	if !ratelimit.Check(c, h.Limiter, ratelimit.LoginPerAccount, ratelimit.AccountKey(c, req.Email)) {
		return
	}

	rlog.Info(c, "login attempt")

	user, err := h.Store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash(), []byte(req.Password))
			rlog.Warn(c, "invalid credentials (user not found)")

			response.Error(c, http.StatusUnauthorized, constants.InvalidCredentials, "Invalid credentials")
			return
		}

		rlog.Error(c, "failed to fetch user", err)

		response.Internal(c)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		rlog.Warn(c, "invalid credentials (password mismatch)", "user_id", user.ID)

		response.Error(c, http.StatusUnauthorized, constants.InvalidCredentials, "Invalid credentials")
		return
	}

	rlog.Info(c, "password verified", "user_id", user.ID)

	// Checked after the password, so only someone who knows it learns that the
	// account exists and isn't verified.
	if !user.EmailVerifiedAt.Valid {
		rlog.Warn(c, "login refused, email not verified", "user_id", user.ID)

		response.Error(c, http.StatusForbidden, constants.EmailNotVerified, "Verify your email before signing in. Check your inbox for the link")
		return
	}

	if err := session.Issue(c, h.Store, h.Config, user); err != nil {
		rlog.Error(c, "failed to issue session", err, "user_id", user.ID)

		response.Internal(c)
		return
	}

	rlog.Info(c, "login successful", "user_id", user.ID)

	response.OK(c, http.StatusOK, "logged in successfully", newUserResponse(user))
}
