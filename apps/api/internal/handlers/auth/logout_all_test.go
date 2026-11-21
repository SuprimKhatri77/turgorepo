package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

// serveAs runs handler as RequireAuth would after accepting userID's token.
func serveAs(handler gin.HandlerFunc, userID string) *httptest.ResponseRecorder {
	r := gin.New()
	r.POST("/", func(c *gin.Context) {
		c.Set(constants.UserIDKey, userID)
		c.Next()
	}, handler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", nil))
	return w
}

func TestLogoutAll(t *testing.T) {
	const userID = "01010101-0101-0101-0101-010101010101"

	t.Run("revokes every session and access token, and clears cookies", func(t *testing.T) {
		repo := &fakeRepo{}
		revoker := &fakeRevoker{}
		before := time.Now()
		w := serveAs(LogoutAll(repo, testConfig(), revoker), userID)

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
		w := serveAs(LogoutAll(&fakeRepo{revokeAllErr: errDB}, testConfig(), revoker), userID)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, revoker.userIDs)
		assert.Empty(t, setCookies(w))
	})

	t.Run("a Redis failure still logs out", func(t *testing.T) {
		repo := &fakeRepo{}
		w := serveAs(LogoutAll(repo, testConfig(), &fakeRevoker{err: errors.New("redis down")}), userID)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, repo.revokedUsers, 1)
		assert.Contains(t, setCookies(w), "refresh_token", "refresh cookie should be cleared")
		assert.Empty(t, setCookies(w)["refresh_token"], "refresh cookie should be cleared")
	})
}
