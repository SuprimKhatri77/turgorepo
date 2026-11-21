package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name     string `json:"name" label:"Name" binding:"required,not_blank,min=2,max=50,person_name" msg_required:"Name is required" msg_min:"Name must be at least 2 characters" msg_max:"Name cannot exceed 50 characters" msg_person_name:"Name can only contain letters, spaces, and ' . -" msg_not_blank:"Name cannot be blank"`
	Email    string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
	Password string `json:"password" trim:"-" label:"Password" binding:"required,not_blank,min=8,max_bytes=72" msg_required:"Password is required" msg_min:"Password must be at least 8 characters" msg_max_bytes:"Password cannot exceed 72 bytes" msg_not_blank:"Password cannot be blank"`
}

// Register creates an account and emails a verification link. It doesn't log
// the user in: login is refused until the address is verified.
func (h *Handler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	req, ok := validator.BindJSON[RegisterRequest](c)
	if !ok {
		return
	}

	req.Email = utils.NormalizeEmail(req.Email)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		rlog.Error(c, "failed to hash password", err)

		response.Internal(c)
		return
	}

	user, err := h.Store.CreateUser(ctx, db.CreateUserParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Role:         constants.RoleMember,
	})

	if err != nil {
		if database.IsUniqueViolation(err) {
			rlog.Warn(c, "user already exists", "email", req.Email)

			response.Error(c, http.StatusConflict, constants.UserAlreadyExists, "User already exists")
			return
		}

		rlog.Error(c, "failed to create user", err)

		response.Internal(c)
		return
	}

	rlog.Info(c, "registration successful", "user_id", user.ID)

	// If sending fails, the user can ask for a new link from the login page.
	h.sendInBackground(c, "verification email", func(ctx context.Context) error {
		return h.sendVerificationEmail(ctx, user)
	})

	response.OK(c, http.StatusCreated, "Check your inbox for a link to verify your email", nil)
}
