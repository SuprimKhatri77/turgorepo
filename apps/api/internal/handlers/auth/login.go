package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
	Password string `json:"password" label:"Password" binding:"required,min=8,max=50" msg_required:"Password is required" msg_min:"Password must be at least 8 characters" msg_max:"Password cannot exceed 50 characters"`
}

func Login(
	queries repository.AuthRepository,
	cfg *config.Config,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		req, ok := validator.BindJSON[LoginRequest](c)
		if !ok {
			return
		}

		utils.TrimStruct(req, "Password")

		rlog.Info(c, "login attempt")

		user, err := queries.GetUserByEmail(ctx, req.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				rlog.Warn(c, "invalid credentials (user not found)")

				c.JSON(http.StatusUnauthorized, types.APIResponse{
					Success: false,
					Message: "Invalid credentials",
					Code:    constants.InvalidCredentials,
				})
				return
			}

			rlog.Error(c, "failed to fetch user", err)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(user.PasswordHash),
			[]byte(req.Password),
		)
		if err != nil {
			rlog.Warn(c, "invalid credentials (password mismatch)", "user_id", user.ID)

			c.JSON(http.StatusUnauthorized, types.APIResponse{
				Success: false,
				Message: "Invalid credentials",
				Code:    constants.InvalidCredentials,
			})
			return
		}

		rlog.Info(c, "password verified", "user_id", user.ID)

		if err := session.IssueSession(c, queries, cfg, user); err != nil {
			rlog.Error(c, "failed to issue session", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "login successful", "user_id", user.ID)

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "logged in successfully",
			Data:    session.PublicUser(user),
		})
	}
}
