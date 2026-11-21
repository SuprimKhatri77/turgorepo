package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"golang.org/x/crypto/bcrypt"
)

func TestResetPassword(t *testing.T) {
	body := func(token, password string) string {
		return `{"token":"` + token + `","password":"` + password + `"}`
	}

	t.Run("sets the password, clears cookies and the login limit, and notifies", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		limiter := &fakeLimiter{}
		mailer := &fakeMailer{}
		w := serve(ResetPassword(resetDeps(repo, limiter, mailer)), body("the-token", "new-password-123"))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, session.HashToken("the-token"), repo.resetParams.TokenHash, "didn't look up the token by hash")
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.resetParams.PasswordHash), []byte("new-password-123")), "stored hash doesn't match the new password")
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			assert.Contains(t, setCookies(w), name, "cookie %s should be cleared", name)
			assert.Empty(t, setCookies(w)[name], "cookie %s should be cleared", name)
		}
		if assert.Len(t, limiter.cleared, 1) {
			assert.True(t, strings.HasPrefix(limiter.cleared[0], "ratelimit:login:account:"), "cleared = %v", limiter.cleared)
		}
		if assert.Len(t, mailer.sent, 1) {
			assert.Contains(t, mailer.sent[0].Subject, "changed")
		}
	})

	t.Run("revokes the user's access tokens", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		deps := resetDeps(repo, &fakeLimiter{}, &fakeMailer{})
		revoker := &fakeRevoker{}
		deps.Revoker = revoker

		before := time.Now()
		serve(ResetPassword(deps), body("the-token", "new-password-123"))

		require.Equal(t, []string{"01010101-0101-0101-0101-010101010101"}, revoker.userIDs)
		assert.WithinRange(t, revoker.cutoff, before, time.Now())
		assert.Equal(t, deps.Config.AccessTokenTTL, revoker.ttl)
	})

	t.Run("a failed revocation doesn't fail the reset", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		deps := resetDeps(repo, &fakeLimiter{}, &fakeMailer{})
		deps.Revoker = &fakeRevoker{err: errors.New("redis down")}
		assert.Equal(t, http.StatusOK, serve(ResetPassword(deps), body("the-token", "new-password-123")).Code)
	})

	t.Run("an invalid token revokes nothing", func(t *testing.T) {
		deps := resetDeps(&resetRepo{resetErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{})
		revoker := &fakeRevoker{}
		deps.Revoker = revoker
		serve(ResetPassword(deps), body("bad", "new-password-123"))
		assert.Empty(t, revoker.userIDs)
	})

	t.Run("a failed notification doesn't fail the reset", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		w := serve(ResetPassword(resetDeps(repo, &fakeLimiter{}, &fakeMailer{err: errors.New("smtp down")})), body("the-token", "new-password-123"))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid or used token", func(t *testing.T) {
		mailer := &fakeMailer{}
		w := serve(ResetPassword(resetDeps(&resetRepo{resetErr: pgx.ErrNoRows}, &fakeLimiter{}, mailer)), body("bad", "new-password-123"))

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, constants.InvalidResetToken, decode(t, w).Code)
		assert.Empty(t, setCookies(w), "a bad link must not log the user out")
		assert.Empty(t, mailer.sent, "sent an email for a failed reset")
	})

	t.Run("database error isn't reported as a bad link", func(t *testing.T) {
		w := serve(ResetPassword(resetDeps(&resetRepo{resetErr: errDB}, &fakeLimiter{}, &fakeMailer{})), body("x", "new-password-123"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("password rules match registration", func(t *testing.T) {
		tooLong := strings.Repeat("😀", 50)
		for _, password := range []string{"short", "", "        ", tooLong} {
			repo := &resetRepo{}
			w := serve(ResetPassword(resetDeps(repo, &fakeLimiter{}, &fakeMailer{})), body("the-token", password))
			assert.Equal(t, http.StatusBadRequest, w.Code, "%q", password)
			assert.Empty(t, repo.resetParams.TokenHash, "%q: token was used despite the invalid password", password)
		}
	})

	t.Run("oversized token is rejected", func(t *testing.T) {
		w := serve(ResetPassword(resetDeps(&resetRepo{}, &fakeLimiter{}, &fakeMailer{})), body(strings.Repeat("a", 129), "new-password-123"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
