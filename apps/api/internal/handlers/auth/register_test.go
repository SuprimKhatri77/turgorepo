package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	body := func(name, email, password string) string {
		b, err := json.Marshal(map[string]string{"name": name, "email": email, "password": password})
		require.NoError(t, err)
		return string(b)
	}

	t.Run("accepts real names", func(t *testing.T) {
		for _, name := range []string{"José García", "O'Brien", "O’Brien", "Mary-Jane", "सुप्रिम", "सुप्रिम खत्री", "J. R. Smith", "Martin Jr."} {
			w := serve(newHandler(&fakeRepo{}).Register, body(name, "a@example.com", testPassword))
			assert.Equal(t, http.StatusCreated, w.Code, "%q: %s", name, w.Body)
		}
	})

	t.Run("rejects names that are not names", func(t *testing.T) {
		for _, name := range []string{"12345", "<script>", "John\nDoe", "--", "   "} {
			w := serve(newHandler(&fakeRepo{}).Register, body(name, "a@example.com", testPassword))
			assert.Equal(t, http.StatusBadRequest, w.Code, "%q", name)
		}
	})

	t.Run("stores the normalized email", func(t *testing.T) {
		repo := &fakeRepo{}
		serve(newHandler(repo).Register, body("Ada", "Ada.L@Example.COM", testPassword))
		if assert.NotNil(t, repo.created) {
			assert.Equal(t, "ada.l@example.com", repo.created.Email)
		}
	})

	t.Run("password over 72 bytes is a validation error", func(t *testing.T) {
		emoji := strings.Repeat("😀", 50) // 50 characters, 200 bytes
		w := serve(newHandler(&fakeRepo{}).Register, body("Ada", "a@example.com", emoji))
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		errs := decode(t, w).Errors
		if assert.Len(t, errs, 1) {
			assert.Equal(t, "password", errs[0].Field)
		}
	})

	t.Run("password of exactly 72 bytes works", func(t *testing.T) {
		w := serve(newHandler(&fakeRepo{}).Register, body("Ada", "a@example.com", strings.Repeat("a", 72)))
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	})

	t.Run("duplicate email is a conflict", func(t *testing.T) {
		repo := &fakeRepo{createErr: &pgconn.PgError{Code: "23505"}}
		w := serve(newHandler(repo).Register, body("Ada", "a@example.com", testPassword))
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("emails a verification link and doesn't log the user in", func(t *testing.T) {
		repo := &fakeRepo{}
		mailer := &fakeMailer{}
		w := serve(newHandler(repo, mailer).Register, body("Ada", "Ada@Example.com", testPassword))

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		assert.Empty(t, setCookies(w))
		assert.Zero(t, repo.sessionsNew)
		assert.Nil(t, decode(t, w).Data, "response should not include the user")
		require.Len(t, mailer.sent, 1)
		require.Equal(t, "ada@example.com", mailer.sent[0].To)
		token := verifyTokenFromEmail(t, mailer.sent[0])
		stored := repo.verificationToken
		require.NotNil(t, stored)
		require.Equal(t, session.HashToken(token), stored.TokenHash)
		require.Equal(t, "ada@example.com", stored.Email)
		assert.WithinRange(t, stored.ExpiresAt.Time, time.Now().Add(23*time.Hour), time.Now().Add(24*time.Hour))
	})

	t.Run("a failed email doesn't fail registration", func(t *testing.T) {
		w := serve(newHandler(&fakeRepo{}, &fakeMailer{err: errors.New("smtp down")}).Register, body("Ada", "a@example.com", testPassword))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("new users are always members", func(t *testing.T) {
		repo := &fakeRepo{}
		serve(newHandler(repo).Register, `{"name":"Ada","email":"a@example.com","password":"`+testPassword+`","role":"admin"}`)
		if assert.NotNil(t, repo.created) {
			assert.Equal(t, constants.RoleMember, repo.created.Role)
		}
	})

	t.Run("invalid fields are rejected and nothing is created", func(t *testing.T) {
		tests := []struct {
			name, body, wantField string
		}{
			{"missing name", `{"email":"a@example.com","password":"` + testPassword + `"}`, "name"},
			{"name too short", body("A", "a@example.com", testPassword), "name"},
			{"name too long", body(strings.Repeat("a", 51), "a@example.com", testPassword), "name"},
			{"missing email", `{"name":"Ada","password":"` + testPassword + `"}`, "email"},
			{"invalid email", body("Ada", "not-an-email", testPassword), "email"},
			{"missing password", `{"name":"Ada","email":"a@example.com"}`, "password"},
			{"password under 8 characters", body("Ada", "a@example.com", "short"), "password"},
			{"blank password", body("Ada", "a@example.com", strings.Repeat(" ", 8)), "password"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := &fakeRepo{}
				mailer := &fakeMailer{}
				w := serve(newHandler(repo, mailer).Register, tt.body)

				require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				res := decode(t, w)
				assert.Equal(t, constants.ValidationFailed, res.Code)
				if assert.Len(t, res.Errors, 1) {
					assert.Equal(t, tt.wantField, res.Errors[0].Field)
				}
				assert.Nil(t, repo.created, "user was created")
				assert.Empty(t, mailer.sent)
			})
		}
	})

	t.Run("malformed JSON is a validation error", func(t *testing.T) {
		repo := &fakeRepo{}
		w := serve(newHandler(repo).Register, `{"name":`)
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, constants.ValidationFailed, decode(t, w).Code)
		assert.Nil(t, repo.created)
	})

	t.Run("stores a bcrypt hash of the untrimmed password", func(t *testing.T) {
		repo := &fakeRepo{}
		password := "  spaces around  "
		serve(newHandler(repo).Register, body("Ada", "a@example.com", password))

		require.NotNil(t, repo.created)
		assert.NotEqual(t, password, repo.created.PasswordHash)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.created.PasswordHash), []byte(password)))
	})

	t.Run("other database errors are a 500 and send no email", func(t *testing.T) {
		for _, err := range []error{errDB, &pgconn.PgError{Code: "23503"}} {
			mailer := &fakeMailer{}
			w := serve(newHandler(&fakeRepo{createErr: err}, mailer).Register, body("Ada", "a@example.com", testPassword))
			assert.Equal(t, http.StatusInternalServerError, w.Code, "%v", err)
			assert.Equal(t, constants.InternalServerError, decode(t, w).Code, "%v", err)
			assert.Empty(t, mailer.sent, "%v", err)
		}
	})

	t.Run("trims the name and email before validating them", func(t *testing.T) {
		repo := &fakeRepo{}
		w := serve(newHandler(repo).Register, body("  Ada Lovelace ", " Ada@Example.com  ", testPassword))

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		require.NotNil(t, repo.created)
		assert.Equal(t, "Ada Lovelace", repo.created.Name)
		assert.Equal(t, "ada@example.com", repo.created.Email)
	})
}
