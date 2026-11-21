package auth

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutAll(t *testing.T) {
	const userID = "01010101-0101-0101-0101-010101010101"

	t.Run("revokes every session and access token, and clears cookies", func(t *testing.T) {
		repo := &fakeRepo{}
		revoker := &fakeRevoker{}
		before := time.Now()
		w := serveAs(newHandler(repo, revoker).LogoutAll, userID)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, []pgtype.UUID{testUUID(1)}, repo.revokedUsers)
		assert.Equal(t, []string{userID}, revoker.userIDs)
		assert.WithinRange(t, revoker.cutoff, before, time.Now())
		assert.Equal(t, testConfig().AccessTokenTTL, revoker.ttl)
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			assert.Contains(t, setCookies(w), name, "cookie %s should be cleared", name)
			assert.Empty(t, setCookies(w)[name], "cookie %s should be cleared", name)
		}
	})

	t.Run("a database error changes nothing", func(t *testing.T) {
		revoker := &fakeRevoker{}
		w := serveAs(newHandler(&fakeRepo{revokeAllErr: errDB}, revoker).LogoutAll, userID)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, revoker.userIDs)
		assert.Empty(t, setCookies(w))
	})

	t.Run("a Redis failure still logs out", func(t *testing.T) {
		repo := &fakeRepo{}
		w := serveAs(newHandler(repo, &fakeRevoker{err: errors.New("redis down")}).LogoutAll, userID)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, repo.revokedUsers, 1)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
		assert.Empty(t, setCookies(w)["refresh_token"], "refresh cookie should be cleared")
	})
}
