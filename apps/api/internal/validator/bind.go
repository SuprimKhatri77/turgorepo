package validator

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
)

// BindJSON decodes the request body into T, trims its string fields (see
// trimStrings) and validates it. Trimming comes first so " ada@example.com "
// passes the email rule instead of being rejected for its spaces.
// On failure it writes a 400 response and returns (nil, false).
func BindJSON[T any](c *gin.Context) (*T, bool) {
	var body T
	err := json.NewDecoder(c.Request.Body).Decode(&body)
	if err == nil {
		trimStrings(&body)
		err = binding.Validator.ValidateStruct(&body)
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			rlog.Warn(c, "request body too large", "limit", tooLarge.Limit)
			response.Error(c, http.StatusRequestEntityTooLarge, constants.PayloadTooLarge, "Request body is too large")
			return nil, false
		}

		rlog.Warn(c, "invalid request payload", "error", err)

		c.JSON(http.StatusBadRequest, response.Body{
			Success: false,
			Message: "Invalid request body",
			Code:    constants.ValidationFailed,
			Errors:  Parse(err, body),
		})
		return nil, false
	}

	return &body, true
}
