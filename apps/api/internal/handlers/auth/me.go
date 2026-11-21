package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

func Me(queries Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userIDFromContext := c.MustGet(constants.UserIDKey).(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			rlog.Error(c, "invalid user_id in context", err, "user_id", userIDFromContext)

			response.Error(c, http.StatusBadRequest, constants.ValidationFailed, "Invalid user ID format")
			return
		}

		user, err := queries.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				rlog.Warn(c, "user not found", "user_id", userID)

				response.Error(c, http.StatusNotFound, constants.UserNotFound, "User not found")
				return
			}

			rlog.Error(c, "failed to fetch user", err, "user_id", userID)

			response.Internal(c)
			return
		}

		rlog.Info(c, "fetched current user", "user_id", userID, "role", user.Role)

		response.OK(c, http.StatusOK, "Valid session", newUserResponse(user))
	}
}
