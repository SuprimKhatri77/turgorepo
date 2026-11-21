package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
)

func (h *Handler) Me(c *gin.Context) {
	ctx := c.Request.Context()

	_, userID, ok := currentUserID(c)
	if !ok {
		return
	}

	user, err := h.Store.GetUserByID(ctx, userID)
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
