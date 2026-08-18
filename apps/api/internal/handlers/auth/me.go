package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Me(queries repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userIDFromContext := c.MustGet(constants.UserIDKey).(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			rlog.Error(c, "invalid user_id in context", err, "user_id", userIDFromContext)

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
				rlog.Warn(c, "user not found", "user_id", userID)

				c.JSON(http.StatusNotFound, types.APIResponse{
					Success: false,
					Message: "User not found",
					Code:    constants.UserNotFound,
				})
				return
			}

			rlog.Error(c, "failed to fetch user", err, "user_id", userID)

			c.JSON(http.StatusInternalServerError, types.APIResponse{
				Success: false,
				Message: "Failed to fetch user",
				Code:    constants.InternalServerError,
			})
			return
		}

		rlog.Info(c, "fetched current user", "user_id", userID, "role", user.Role)

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Valid session",
			Data:    session.PublicUser(user),
		})
	}
}
