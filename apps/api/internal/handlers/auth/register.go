package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
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
	Name     string `json:"name" label:"Name" binding:"required,not_blank,min=2,max=50,alphaspace" msg_required:"Name is required" msg_min:"Name must be at least 2 characters" msg_max:"Name cannot exceed 50 characters" msg_alphaspace:"Name can only contain letters and spaces" msg_not_blank:"Name cannot be blank"`
	Email    string `json:"email" label:"Email" binding:"required,email" msg_required:"Email is required" msg_email:"Enter a valid email address"`
	Password string `json:"password" label:"Password" binding:"required,not_blank,min=8,max=50" msg_required:"Password is required" msg_min:"Password must be at least 8 characters" msg_max:"Password cannot exceed 50 characters" msg_not_blank:"Password cannot be blank"`
}

func Register(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		req, ok := validator.BindJSON[RegisterRequest](c)
		if !ok {
			return
		}

		utils.TrimStruct(req, "Password")

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
			Role:         constants.RoleMember,
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

		if err := session.IssueSession(c, queries, cfg, user); err != nil {
			rlog.Error(c, "failed to issue session", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Something went wrong",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "registration successful", "user_id", user.ID)

		c.JSON(http.StatusCreated, types.APIResponse{
			Success: true,
			Message: "Registration successful",
			Data:    session.PublicUser(user),
		})
	}
}
