package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

func TestVerifyEmail(t *testing.T) {
	body := func(token string) string { return `{"token":"` + token + `"}` }

	t.Run("verifies by the token's hash without logging in", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(VerifyEmail(verifyDeps(repo, &fakeLimiter{}, &fakeMailer{})), body("the-token"))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, session.HashToken("the-token"), repo.verifiedHash, "didn't look up the token by hash")
		assert.Empty(t, setCookies(w), "verifying must not log the user in")
		assert.Zero(t, repo.sessionsNew, "verifying must not log the user in")
	})

	t.Run("invalid, used, expired or outdated token", func(t *testing.T) {
		w := serve(VerifyEmail(verifyDeps(&fakeRepo{verifyErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{})), body("bad"))
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, constants.InvalidVerificationToken, decode(t, w).Code)
	})

	t.Run("database error isn't reported as a bad link", func(t *testing.T) {
		w := serve(VerifyEmail(verifyDeps(&fakeRepo{verifyErr: errDB}, &fakeLimiter{}, &fakeMailer{})), body("x"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("missing or oversized token is rejected before the database", func(t *testing.T) {
		for _, b := range []string{`{}`, body(strings.Repeat("a", 129))} {
			repo := &fakeRepo{}
			w := serve(VerifyEmail(verifyDeps(repo, &fakeLimiter{}, &fakeMailer{})), b)
			assert.Equal(t, http.StatusBadRequest, w.Code, b)
			assert.Empty(t, repo.verifiedHash, "%s: looked up", b)
		}
	})
}

func TestResendVerification(t *testing.T) {
	const body = `{"email":"Ada@Example.com"}`

	t.Run("sends a new link to an unverified account", func(t *testing.T) {
		repo := &fakeRepo{user: unverifiedUser(t)}
		mailer := &fakeMailer{}
		w := serve(ResendVerification(verifyDeps(repo, &fakeLimiter{}, mailer)), body)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ada@example.com", repo.lookedUp)
		require.Len(t, mailer.sent, 1)
		token := verifyTokenFromEmail(t, mailer.sent[0])
		assert.Equal(t, session.HashToken(token), repo.verificationToken.TokenHash, "stored hash doesn't match the emailed token")
	})

	t.Run("responds the same however it goes", func(t *testing.T) {
		tests := []struct {
			name      string
			repo      *fakeRepo
			limiter   *fakeLimiter
			mailer    *fakeMailer
			wantEmail bool
		}{
			{"unverified account", &fakeRepo{user: unverifiedUser(t)}, &fakeLimiter{}, &fakeMailer{}, true},
			{"already verified", &fakeRepo{user: testUser(t)}, &fakeLimiter{}, &fakeMailer{}, false},
			{"unknown email", &fakeRepo{getUserErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{}, false},
			{"database error", &fakeRepo{getUserErr: errDB}, &fakeLimiter{}, &fakeMailer{}, false},
			{"token not stored", &fakeRepo{user: unverifiedUser(t), verificationErr: errDB}, &fakeLimiter{}, &fakeMailer{}, false},
			{"email fails", &fakeRepo{user: unverifiedUser(t)}, &fakeLimiter{}, &fakeMailer{err: errors.New("smtp down")}, true},
			{"email limit reached", &fakeRepo{user: unverifiedUser(t)}, &fakeLimiter{blocked: true}, &fakeMailer{}, false},
		}

		var firstBody string
		for _, tt := range tests {
			w := serve(ResendVerification(verifyDeps(tt.repo, tt.limiter, tt.mailer)), body)
			assert.Equal(t, http.StatusOK, w.Code, tt.name)
			if firstBody == "" {
				firstBody = w.Body.String()
			}
			assert.Equal(t, firstBody, w.Body.String(), "%s: body differs", tt.name)
			assert.Equal(t, tt.wantEmail, len(tt.mailer.sent) > 0, "%s: emailed", tt.name)
		}
	})

	t.Run("email limit applies across IPs, without the raw address", func(t *testing.T) {
		limiter := &fakeLimiter{}
		serve(ResendVerification(verifyDeps(&fakeRepo{user: unverifiedUser(t)}, limiter, &fakeMailer{})), body)
		require.Len(t, limiter.keys, 1)
		require.True(t, strings.HasPrefix(limiter.keys[0], "ratelimit:resend-verification:email:"), "key = %s", limiter.keys[0])
		assert.NotContains(t, limiter.keys[0], "ada@example.com", "key has the address")
	})

	t.Run("invalid email is a validation error", func(t *testing.T) {
		w := serve(ResendVerification(verifyDeps(&fakeRepo{}, &fakeLimiter{}, &fakeMailer{})), `{"email":"nope"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
