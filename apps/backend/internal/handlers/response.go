package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/backend/internal/errors"
)

// JSON responds with JSON and the given status code.
func JSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

// JSONError responds with an error using AppError.
func JSONError(c *gin.Context, err error) {
	code := errors.HTTPStatus(err)
	msg := errors.ResponseMessage(err)
	c.JSON(code, gin.H{"error": msg})
}

// BindJSON binds the request body to v and returns true on success; on failure it responds with 400 and returns false.
func BindJSON(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		JSONError(c, errors.WithMessage(errors.ErrBadRequest, "invalid JSON: "+err.Error()))
		return false
	}
	return true
}
