// Package response writes the JSON envelope every API response uses.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

// Body is the envelope of every API response. It matches ApiResponse in
// @repo/types.
type Body struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Code    string       `json:"code,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
	Data    any          `json:"data,omitempty"`
	Meta    any          `json:"meta,omitempty"`
}

// FieldError describes one invalid field in a request body.
type FieldError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// PaginationMeta goes in Body.Meta for list endpoints.
type PaginationMeta struct {
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}

// OK writes a successful response. data may be nil.
func OK(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Body{Success: true, Message: message, Data: data})
}

// Error writes a failed response. code is one of the constants error codes,
// which clients branch on; message is for people.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Body{Success: false, Message: message, Code: code})
}

// Abort is Error for middleware: it also stops the remaining handlers.
func Abort(c *gin.Context, status int, code, message string) {
	Error(c, status, code, message)
	c.Abort()
}

// Internal writes the 500 for an unexpected failure. Log the cause first: the
// response deliberately says nothing about it.
func Internal(c *gin.Context) {
	Error(c, http.StatusInternalServerError, constants.InternalServerError, "Something went wrong")
}
