package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

func TestRefresh(t *testing.T) {
	t.Run("current token rotates", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

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
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

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
		w := serve(newHandler(repo, revoker).Refresh, "", refreshCookie(token))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Len(t, repo.revoked, 1)
		// Whoever holds a stolen refresh token loses its access tokens too.
		assert.Equal(t, []string{testSessionID}, revoker.sessions, "revoked session tokens")
	})

	t.Run("expired or revoked session clears cookies", func(t *testing.T) {
		user, token, _ := refreshFixture(t)
		w := serve(newHandler(&fakeRepo{user: user, sessionErr: pgx.ErrNoRows}).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
		assert.Empty(t, setCookies(w)["refresh_token"], "refresh cookie should be cleared")
	})

	t.Run("database outage keeps the user logged in", func(t *testing.T) {
		user, token, _ := refreshFixture(t)
		w := serve(newHandler(&fakeRepo{user: user, sessionErr: errDB}).Refresh, "", refreshCookie(token))

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
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.NotContains(t, setCookies(w), "refresh_token", "the losing request must not set a refresh token")
	})

	t.Run("a non-UUID family ID clears cookies", func(t *testing.T) {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, session.RefreshClaims{
			UserID:           "01010101-0101-0101-0101-010101010101",
			FamilyID:         "not-a-uuid",
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
		}).SignedString([]byte(testConfig().JWTRefreshSecret))
		require.NoError(t, err)
		w := serve(newHandler(&fakeRepo{}).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, constants.InvalidToken, decode(t, w).Code)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
	})

	t.Run("a failed user lookup is a 500 and keeps the cookies", func(t *testing.T) {
		_, token, sess := refreshFixture(t)
		repo := &fakeRepo{getUserErr: errDB, sessions: []db.Session{sess}}
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.False(t, repo.rotated, "rotated")
		assert.Empty(t, setCookies(w))
	})

	t.Run("a failed rotation is a 500 and keeps the cookies", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}, rotateErr: errDB}
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, setCookies(w), "the old tokens must stay usable")
	})

	t.Run("losing the rotation race to a logout clears cookies", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		// The reload finds no session: the winner was a logout, not a refresh.
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}, rotateErr: pgx.ErrNoRows}
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, constants.TokenInvalid, decode(t, w).Code)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
		assert.Empty(t, setCookies(w)["refresh_token"], "refresh cookie should be cleared")
	})

	t.Run("losing the rotation race outside the grace window clears cookies", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		reloaded := sess
		reloaded.PreviousTokenHash = pgtype.Text{String: "hash-of-another-token", Valid: true}
		reloaded.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		reloaded.CurrentTokenHash = "hash-of-winner"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess, reloaded}, rotateErr: pgx.ErrNoRows}
		w := serve(newHandler(repo).Refresh, "", refreshCookie(token))

		require.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, constants.TokenInvalid, decode(t, w).Code)
		assert.NotEmpty(t, setCookies(w), "cookies should be cleared")
		assert.Empty(t, setCookies(w)["access_token"], "no new access token")
	})

	t.Run("missing and invalid cookies are rejected", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, serve(newHandler(&fakeRepo{}).Refresh, "").Code, "missing")
		assert.Equal(t, http.StatusUnauthorized, serve(newHandler(&fakeRepo{}).Refresh, "", refreshCookie("not-a-jwt")).Code, "invalid")
	})
}

func TestWithinGrace(t *testing.T) {
	now := time.Now()
	window := 5 * time.Minute
	rotated := func(ago time.Duration) db.Session {
		return db.Session{
			PreviousTokenHash: pgtype.Text{String: "previous", Valid: true},
			PreviousRotatedAt: pgtype.Timestamptz{Time: now.Add(-ago), Valid: true},
		}
	}

	tests := []struct {
		name string
		sess db.Session
		hash string
		want bool
	}{
		{"previous token just rotated", rotated(time.Second), "previous", true},
		{"previous token just inside the window", rotated(window - time.Nanosecond), "previous", true},
		{"previous token at the window's end", rotated(window), "previous", false},
		{"previous token after the window", rotated(window + time.Minute), "previous", false},
		{"a different token", rotated(time.Second), "other", false},
		{"never rotated", db.Session{}, "", false},
		{"no rotation time", db.Session{PreviousTokenHash: pgtype.Text{String: "previous", Valid: true}}, "previous", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, withinGrace(tt.sess, tt.hash, window, now))
		})
	}
}
