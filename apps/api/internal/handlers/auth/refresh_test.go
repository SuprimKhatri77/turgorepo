package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

func TestRefresh(t *testing.T) {
	t.Run("current token rotates", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		w := serve(Refresh(repo, testConfig(), &fakeRevoker{}), "", refreshCookie(token))

		require.Equal(t, http.StatusOK, w.Code)
		require.True(t, repo.rotated, "rotated")
		cookies := setCookies(w)
		assert.NotEmpty(t, cookies["access_token"], "access token not set")
		assert.NotEmpty(t, cookies["refresh_token"], "refresh token not set")
	})

	t.Run("previous token in the grace window only gets an access token", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		sess.PreviousTokenHash = pgtype.Text{String: sess.CurrentTokenHash, Valid: true}
		sess.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}
		sess.CurrentTokenHash = "hash-of-newer-token"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		w := serve(Refresh(repo, testConfig(), &fakeRevoker{}), "", refreshCookie(token))

		require.Equal(t, http.StatusOK, w.Code)
		require.False(t, repo.rotated, "rotated")
		cookies := setCookies(w)
		assert.NotEmpty(t, cookies["access_token"], "access token not set")
		assert.NotContains(t, cookies, "refresh_token", "refresh token should not change")
	})

	t.Run("previous token after the grace window revokes the session", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		sess.PreviousTokenHash = pgtype.Text{String: sess.CurrentTokenHash, Valid: true}
		sess.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true}
		sess.CurrentTokenHash = "hash-of-newer-token"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		revoker := &fakeRevoker{}
		w := serve(Refresh(repo, testConfig(), revoker), "", refreshCookie(token))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Len(t, repo.revoked, 1)
		// Whoever holds a stolen refresh token loses its access tokens too.
		assert.Equal(t, []string{testSessionID}, revoker.sessions, "revoked session tokens")
	})

	t.Run("expired or revoked session clears cookies", func(t *testing.T) {
		user, token, _ := refreshFixture(t)
		w := serve(Refresh(&fakeRepo{user: user, sessionErr: pgx.ErrNoRows}, testConfig(), &fakeRevoker{}), "", refreshCookie(token))

		require.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
		assert.Empty(t, setCookies(w)["refresh_token"], "refresh cookie should be cleared")
	})

	t.Run("database outage keeps the user logged in", func(t *testing.T) {
		user, token, _ := refreshFixture(t)
		w := serve(Refresh(&fakeRepo{user: user, sessionErr: errDB}, testConfig(), &fakeRevoker{}), "", refreshCookie(token))

		require.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Empty(t, setCookies(w), "cookies should be untouched")
		assert.Equal(t, constants.ServiceUnavailable, decode(t, w).Code)
	})

	t.Run("losing the rotation race falls back to the grace path", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		reloaded := sess
		reloaded.PreviousTokenHash = pgtype.Text{String: sess.CurrentTokenHash, Valid: true}
		reloaded.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		reloaded.CurrentTokenHash = "hash-of-winner"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess, reloaded}, rotateErr: pgx.ErrNoRows}
		w := serve(Refresh(repo, testConfig(), &fakeRevoker{}), "", refreshCookie(token))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.NotContains(t, setCookies(w), "refresh_token", "the losing request must not set a refresh token")
	})

	t.Run("missing and invalid cookies are rejected", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, serve(Refresh(&fakeRepo{}, testConfig(), &fakeRevoker{}), "").Code, "missing")
		assert.Equal(t, http.StatusUnauthorized, serve(Refresh(&fakeRepo{}, testConfig(), &fakeRevoker{}), "", refreshCookie("not-a-jwt")).Code, "invalid")
	})
}
