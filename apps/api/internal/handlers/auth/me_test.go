package auth

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

func TestMe(t *testing.T) {
	const userID = "01010101-0101-0101-0101-010101010101"

	t.Run("returns the signed-in user's public fields", func(t *testing.T) {
		w := serveAs(newHandler(&fakeRepo{user: testUser(t)}).Me, userID)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var res struct {
			Data map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.Equal(t, userID, res.Data["id"])
		assert.Equal(t, "ada@example.com", res.Data["email"])
		assert.NotContains(t, res.Data, "password_hash")
	})

	t.Run("a deleted user is a 404", func(t *testing.T) {
		w := serveAs(newHandler(&fakeRepo{getUserErr: pgx.ErrNoRows}).Me, userID)
		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, constants.UserNotFound, decode(t, w).Code)
	})

	t.Run("a database error is a 500", func(t *testing.T) {
		w := serveAs(newHandler(&fakeRepo{getUserErr: errDB}).Me, userID)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, constants.InternalServerError, decode(t, w).Code)
	})

	t.Run("a malformed user ID from the middleware is a server error", func(t *testing.T) {
		w := serveAs(newHandler(&fakeRepo{user: testUser(t)}).Me, "not-a-uuid")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
