package errors

import (
	"errors"
	"net/http"
)

// AppError represents an application error with HTTP status.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates an AppError.
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Common constructors.
var (
	ErrNotFound     = NewAppError(http.StatusNotFound, "resource not found", nil)
	ErrBadRequest   = NewAppError(http.StatusBadRequest, "bad request", nil)
	ErrInternal     = NewAppError(http.StatusInternalServerError, "internal server error", nil)
)

// WithMessage returns a copy of the error with a new message.
func WithMessage(ae *AppError, msg string) *AppError {
	return NewAppError(ae.Code, msg, ae.Err)
}

// Wrap wraps an underlying error with an AppError.
func Wrap(ae *AppError, err error) *AppError {
	return NewAppError(ae.Code, ae.Message, err)
}

// HTTPStatus returns the HTTP status code for an error.
func HTTPStatus(err error) int {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code
	}
	return http.StatusInternalServerError
}

// ResponseMessage returns a safe message for API response.
func ResponseMessage(err error) string {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Message
	}
	return "internal server error"
}
