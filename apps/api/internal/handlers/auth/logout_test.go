package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

func TestLogout(t *testing.T) {
	t.Run("revokes the session and its access tokens, and clears cookies", func(t *testing.T) {
		_, token, _ := refreshFixture(t)
		repo := &fakeRepo{}
		revoker := &fakeRevoker{}
		w := serve(Logout(repo, testConfig(), revoker), "", refreshCookie(token))

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, []pgtype.UUID{testUUID(2)}, repo.revoked)
		assert.Equal(t, []string{testSessionID}, revoker.sessions, "revoked session tokens")
		assert.Equal(t, testConfig().AccessTokenTTL, revoker.ttl)
		assert.Contains(t, setCookies(w), "access_token", "access cookie should be cleared")
		assert.Empty(t, setCookies(w)["access_token"], "access cookie should be cleared")
	})

	t.Run("falls back to the access token without a refresh token", func(t *testing.T) {
		tokens, err := session.NewTokens(testConfig(), testUser(t), testUUID(2))
		require.NoError(t, err)
		access := &http.Cookie{Name: "access_token", Value: tokens.AccessToken}
		for _, cookies := range [][]*http.Cookie{{access}, {access, refreshCookie("not-a-jwt")}} {
			repo := &fakeRepo{}
			revoker := &fakeRevoker{}
			w := serve(Logout(repo, testConfig(), revoker), "", cookies...)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, []pgtype.UUID{testUUID(2)}, repo.revoked)
			assert.Equal(t, []string{testSessionID}, revoker.sessions, "revoked session tokens")
		}
	})

	t.Run("a Redis failure still logs out", func(t *testing.T) {
		_, token, _ := refreshFixture(t)
		repo := &fakeRepo{}
		w := serve(Logout(repo, testConfig(), &fakeRevoker{err: errors.New("redis down")}), "", refreshCookie(token))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, repo.revoked, 1)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
		assert.Empty(t, setCookies(w)["refresh_token"], "refresh cookie should be cleared")
	})

	t.Run("works without a session", func(t *testing.T) {
		repo := &fakeRepo{}
		revoker := &fakeRevoker{}
		for _, cookies := range [][]*http.Cookie{nil, {refreshCookie("not-a-jwt")}} {
			assert.Equal(t, http.StatusOK, serve(Logout(repo, testConfig(), revoker), "", cookies...).Code)
		}
		assert.Empty(t, repo.revoked)
		assert.Empty(t, revoker.sessions)
	})
}
