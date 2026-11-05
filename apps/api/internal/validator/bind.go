package validator

import (
	"errors"
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
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			rlog.Warn(c, "request body too large", "limit", tooLarge.Limit)
			c.JSON(http.StatusRequestEntityTooLarge, types.APIResponse{
				Success: false,
				Message: "Request body is too large",
				Code:    constants.PayloadTooLarge,
			})
			return nil, false
		}

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
