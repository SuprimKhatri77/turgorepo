package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Me(queries repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userIDFromContext := c.MustGet("userID").(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			slog.Error("invalid user_id in context",
				"user_id", userIDFromContext,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
				"error", err,
			)

			c.JSON(http.StatusBadRequest, types.APIResponse{
				Success: false,
				Message: "Invalid user ID format",
				Code:    constants.ValidationFailed,
			})
			return
		}

		user, err := queries.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.Warn("user not found",
					"user_id", userID,
					"path", c.FullPath(),
					"ip", c.ClientIP(),
				)

				c.JSON(http.StatusNotFound, types.APIResponse{
					Success: false,
					Message: "User not found",
					Code:    constants.UserNotFound,
				})
				return
			}

			slog.Error("failed to fetch user",
				"user_id", userID,
				"error", err,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to fetch user",
				Code:    constants.InternalServerError,
			})
			return
		}

		slog.Info("fetched current user",
			"user_id", userID,
			"role", user.Role,
			"path", c.FullPath(),
			"ip", c.ClientIP(),
		)

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Valid session",
			Data: db.User{
				ID:       user.ID,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
				ImageUrl: user.ImageUrl,
			},
		})
	}
}
