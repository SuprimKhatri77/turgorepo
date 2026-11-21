package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

// These pin the validation error each auth request returns, field by field, so
// a change to a binding tag or a default message shows up here.

// fieldError is the code and message one field failed with.
type fieldError struct{ code, message string }

type validationCase struct {
	name string
	body string
	want map[string]fieldError
}

// checkValidation binds each case's body into T and compares the field errors.
func checkValidation[T any](t *testing.T, cases []validationCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))

			_, ok := validator.BindJSON[T](c)
			require.False(t, ok, "body should be rejected")
			require.Equal(t, http.StatusBadRequest, w.Code)

			got := map[string]fieldError{}
			for _, e := range decode(t, w).Errors {
				got[e.Field] = fieldError{e.Code, e.Message}
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

var (
	validName     = `"name":"Ada Lovelace"`
	validEmail    = `"email":"ada@example.com"`
	validPassword = `"password":"password123"`
	// 37 two-byte characters: long enough for min=8 but 74 bytes, over bcrypt's 72.
	tooManyBytes = strings.Repeat("é", 37)
	longToken    = strings.Repeat("a", 129)
)

var (
	errEmailRequired    = fieldError{"REQUIRED_FIELD", "Email is required"}
	errEmailInvalid     = fieldError{"INVALID_EMAIL", "Enter a valid email address"}
	errPasswordRequired = fieldError{"REQUIRED_FIELD", "Password is required"}
	errPasswordShort    = fieldError{"TOO_SHORT", "Password must be at least 8 characters"}
	errPasswordBlank    = fieldError{"BLANK_FIELD", "Password cannot be blank"}
	errPasswordBytes    = fieldError{"TOO_LONG", "Password is too long"}
)

func TestRegisterRequestValidation(t *testing.T) {
	checkValidation[RegisterRequest](t, []validationCase{
		{"empty body", `{}`, map[string]fieldError{
			"name":     {"REQUIRED_FIELD", "Name is required"},
			"email":    errEmailRequired,
			"password": errPasswordRequired,
		}},
		{"name of only spaces", `{"name":"   ",` + validEmail + `,` + validPassword + `}`, map[string]fieldError{
			"name": {"REQUIRED_FIELD", "Name is required"},
		}},
		{"short name", `{"name":"A",` + validEmail + `,` + validPassword + `}`, map[string]fieldError{
			"name": {"TOO_SHORT", "Name must be at least 2 characters"},
		}},
		{"long name", `{"name":"` + strings.Repeat("a", 51) + `",` + validEmail + `,` + validPassword + `}`, map[string]fieldError{
			"name": {"TOO_LONG", "Name cannot exceed 50 characters"},
		}},
		{"name with digits", `{"name":"R2D2",` + validEmail + `,` + validPassword + `}`, map[string]fieldError{
			"name": {"INVALID_FORMAT", "Name can only contain letters, spaces, and ' . -"},
		}},
		{"bad email", `{` + validName + `,"email":"nope",` + validPassword + `}`, map[string]fieldError{
			"email": errEmailInvalid,
		}},
		{"short password", `{` + validName + `,` + validEmail + `,"password":"short"}`, map[string]fieldError{
			"password": errPasswordShort,
		}},
		{"password of only spaces", `{` + validName + `,` + validEmail + `,"password":"        "}`, map[string]fieldError{
			"password": errPasswordBlank,
		}},
		{"password over 72 bytes", `{` + validName + `,` + validEmail + `,"password":"` + tooManyBytes + `"}`, map[string]fieldError{
			"password": errPasswordBytes,
		}},
	})
}

func TestLoginRequestValidation(t *testing.T) {
	checkValidation[LoginRequest](t, []validationCase{
		{"empty body", `{}`, map[string]fieldError{
			"email":    errEmailRequired,
			"password": errPasswordRequired,
		}},
		{"bad email", `{"email":"nope",` + validPassword + `}`, map[string]fieldError{
			"email": errEmailInvalid,
		}},
		{"password over 72 bytes", `{` + validEmail + `,"password":"` + tooManyBytes + `"}`, map[string]fieldError{
			"password": errPasswordBytes,
		}},
	})
}

func TestForgotPasswordRequestValidation(t *testing.T) {
	checkValidation[ForgotPasswordRequest](t, []validationCase{
		{"empty body", `{}`, map[string]fieldError{"email": errEmailRequired}},
		{"bad email", `{"email":"nope"}`, map[string]fieldError{"email": errEmailInvalid}},
	})
}

func TestResetPasswordRequestValidation(t *testing.T) {
	checkValidation[ResetPasswordRequest](t, []validationCase{
		{"empty body", `{}`, map[string]fieldError{
			"token":    {"REQUIRED_FIELD", "Reset link is missing"},
			"password": errPasswordRequired,
		}},
		{"long token", `{"token":"` + longToken + `",` + validPassword + `}`, map[string]fieldError{
			"token": {"TOO_LONG", "Reset link is invalid"},
		}},
		{"short password", `{"token":"abc","password":"short"}`, map[string]fieldError{
			"password": errPasswordShort,
		}},
		{"password of only spaces", `{"token":"abc","password":"        "}`, map[string]fieldError{
			"password": errPasswordBlank,
		}},
		{"password over 72 bytes", `{"token":"abc","password":"` + tooManyBytes + `"}`, map[string]fieldError{
			"password": errPasswordBytes,
		}},
	})
}

func TestVerifyEmailRequestValidation(t *testing.T) {
	checkValidation[VerifyEmailRequest](t, []validationCase{
		{"empty body", `{}`, map[string]fieldError{
			"token": {"REQUIRED_FIELD", "Verification link is missing"},
		}},
		{"long token", `{"token":"` + longToken + `"}`, map[string]fieldError{
			"token": {"TOO_LONG", "Verification link is invalid"},
		}},
	})
}

func TestResendVerificationRequestValidation(t *testing.T) {
	checkValidation[ResendVerificationRequest](t, []validationCase{
		{"empty body", `{}`, map[string]fieldError{"email": errEmailRequired}},
		{"bad email", `{"email":"nope"}`, map[string]fieldError{"email": errEmailInvalid}},
	})
}
