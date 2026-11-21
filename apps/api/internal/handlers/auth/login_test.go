package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

func TestLogin(t *testing.T) {
	t.Run("sets cookies and returns only public fields", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(newHandler(repo).Login, `{"email":"ada@example.com","password":"`+testPassword+`"}`)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		cookies := setCookies(w)
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			assert.NotEmpty(t, cookies[name], "cookie %s not set", name)
		}

		var res struct {
			Data map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.Equal(t, "01010101-0101-0101-0101-010101010101", res.Data["id"])
		for _, key := range []string{"created_at", "updated_at", "image_url", "password_hash"} {
			assert.NotContains(t, res.Data, key)
		}
	})

	t.Run("refuses an unverified email without creating a session", func(t *testing.T) {
		repo := &fakeRepo{user: unverifiedUser(t)}
		w := serve(newHandler(repo).Login, `{"email":"ada@example.com","password":"`+testPassword+`"}`)

		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.Equal(t, constants.EmailNotVerified, decode(t, w).Code)
		assert.Zero(t, repo.sessionsNew)
		assert.Empty(t, setCookies(w))
	})

	t.Run("an unverified account with the wrong password looks like any wrong password", func(t *testing.T) {
		unverified := serve(newHandler(&fakeRepo{user: unverifiedUser(t)}).Login, `{"email":"ada@example.com","password":"wrong-password"}`)
		verified := serve(newHandler(&fakeRepo{user: testUser(t)}).Login, `{"email":"ada@example.com","password":"wrong-password"}`)
		assert.Equal(t, http.StatusUnauthorized, unverified.Code)
		assert.Equal(t, verified.Body.String(), unverified.Body.String())
	})

	t.Run("looks up the normalized email", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		serve(newHandler(repo).Login, `{"email":"Ada@Example.COM","password":"`+testPassword+`"}`)
		assert.Equal(t, "ada@example.com", repo.lookedUp)
	})

	t.Run("wrong password and unknown email respond the same", func(t *testing.T) {
		wrong := serve(newHandler(&fakeRepo{user: testUser(t)}).Login, `{"email":"ada@example.com","password":"wrong-password"}`)
		unknown := serve(newHandler(&fakeRepo{getUserErr: pgx.ErrNoRows}).Login, `{"email":"no@example.com","password":"wrong-password"}`)

		require.Equal(t, http.StatusUnauthorized, wrong.Code, "wrong password")
		require.Equal(t, http.StatusUnauthorized, unknown.Code, "unknown email")
		assert.Equal(t, wrong.Body.String(), unknown.Body.String())
	})

	t.Run("database error is a 500, not invalid credentials", func(t *testing.T) {
		w := serve(newHandler(&fakeRepo{getUserErr: errDB}).Login, `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("account limit blocks before checking the password", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(newHandler(repo, &fakeLimiter{blocked: true}).Login, `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		require.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Empty(t, repo.lookedUp, "user was looked up despite the limit")
		assert.Equal(t, constants.RateLimited, decode(t, w).Code)
	})

	t.Run("account limit treats email case variants as one account", func(t *testing.T) {
		limiter := &fakeLimiter{}
		user := testUser(t)
		serve(newHandler(&fakeRepo{user: user}, limiter).Login, `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		serve(newHandler(&fakeRepo{user: user}, limiter).Login, `{"email":"ADA@example.com","password":"`+testPassword+`"}`)
		if assert.Len(t, limiter.keys, 2) {
			assert.Equal(t, limiter.keys[0], limiter.keys[1])
		}
		assert.NotContains(t, limiter.keys[0], "ada@example.com", "key contains the raw email")
	})

	t.Run("invalid body is rejected before the limiter and the lookup", func(t *testing.T) {
		tests := []struct {
			name      string
			body      string
			wantField string
		}{
			{"malformed JSON", `{"email":`, ""},
			{"missing email", `{"password":"` + testPassword + `"}`, "email"},
			{"invalid email", `{"email":"not-an-email","password":"` + testPassword + `"}`, "email"},
			{"missing password", `{"email":"ada@example.com"}`, "password"},
			{"password over 72 bytes", `{"email":"ada@example.com","password":"` + strings.Repeat("a", 73) + `"}`, "password"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := &fakeRepo{user: testUser(t)}
				limiter := &fakeLimiter{}
				w := serve(newHandler(repo, limiter).Login, tt.body)

				require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				res := decode(t, w)
				assert.Equal(t, constants.ValidationFailed, res.Code)
				if tt.wantField != "" && assert.Len(t, res.Errors, 1) {
					assert.Equal(t, tt.wantField, res.Errors[0].Field)
				}
				assert.Empty(t, limiter.keys, "limiter was hit")
				assert.Empty(t, repo.lookedUp, "user was looked up")
			})
		}
	})

	t.Run("the password is not trimmed", func(t *testing.T) {
		w := serve(newHandler(&fakeRepo{user: testUser(t)}).Login, `{"email":"ada@example.com","password":" `+testPassword+` "}`)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("a failed session insert is a 500 without cookies", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t), newSessErr: errDB}
		w := serve(newHandler(repo).Login, `{"email":"ada@example.com","password":"`+testPassword+`"}`)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, constants.InternalServerError, decode(t, w).Code)
		assert.Empty(t, setCookies(w))
	})

	t.Run("an email with surrounding spaces logs in", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(newHandler(repo).Login, `{"email":"  ada@example.com ","password":"`+testPassword+`"}`)

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "ada@example.com", repo.lookedUp)
	})
}
