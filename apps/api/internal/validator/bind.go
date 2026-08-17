package validator

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

// BindJSON binds and validates the request body into T.
// On failure it writes a 400 response and returns (nil, false).
func BindJSON[T any](c *gin.Context) (*T, bool) {
	var body T
	if err := c.ShouldBindJSON(&body); err != nil {
		rlog.Warn(c, "invalid request payload", "error", err)

		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Code:    constants.ValidationFailed,
			Errors:  Parse(err, body),
		})
		return nil, false
	}

	return &body, true
}
