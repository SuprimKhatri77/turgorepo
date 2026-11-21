package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

const testAccessSecret = "access-secret"

const testSessionID = "session-1"

type fakeRevoker struct {
	cutoff         time.Time
	found          bool
	revokedSession string
	err            error
	calls          int
	userID         string
	sessionID      string
}

func (r *fakeRevoker) RevokeTokensBefore(context.Context, string, time.Time, time.Duration) error {
	return nil
}

func (r *fakeRevoker) RevokeSession(context.Context, string, time.Duration) error {
	return nil
}

func (r *fakeRevoker) Revocations(_ context.Context, userID, sessionID string) (session.Revocations, error) {
	r.calls++
	r.userID = userID
	r.sessionID = sessionID
	return session.Revocations{
		Cutoff:         r.cutoff,
		HasCutoff:      r.found,
		SessionRevoked: r.revokedSession != "" && r.revokedSession == sessionID,
	}, r.err
}

func signToken(t *testing.T, secret string, issuedAt *jwt.NumericDate) string {
	t.Helper()
	return signSessionToken(t, secret, testSessionID, issuedAt)
}

func signSessionToken(t *testing.T, secret, sessionID string, issuedAt *jwt.NumericDate) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, session.AccessClaims{
		UserID:    "user-1",
		SessionID: sessionID,
		Role:      constants.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  issuedAt,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(secret))
	require.NoError(t, err)
	return token
}

func serveAuthed(revoker session.TokenRevoker, token string) (*httptest.ResponseRecorder, bool) {
	gin.SetMode(gin.TestMode)

	reached := false
	r := gin.New()
	r.GET("/", RequireAuth(&config.Config{JWTAccessSecret: testAccessSecret}, revoker), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, reached
}

func TestRequireAuthRevocation(t *testing.T) {
	cutoff := time.Now().Add(-time.Minute)
	at := func(moment time.Time) *jwt.NumericDate { return jwt.NewNumericDate(moment) }

	tests := []struct {
		name     string
		revoker  *fakeRevoker
		issuedAt *jwt.NumericDate
		want     int
	}{
		{"before the cutoff", &fakeRevoker{cutoff: cutoff, found: true}, at(cutoff.Add(-time.Minute)), http.StatusUnauthorized},
		{"same second as the cutoff", &fakeRevoker{cutoff: cutoff, found: true}, at(cutoff), http.StatusUnauthorized},
		{"after the cutoff", &fakeRevoker{cutoff: cutoff, found: true}, at(cutoff.Add(time.Second)), http.StatusOK},
		{"no cutoff", &fakeRevoker{}, at(cutoff), http.StatusOK},
		{"no issue time with a cutoff", &fakeRevoker{cutoff: cutoff, found: true}, nil, http.StatusUnauthorized},
		{"no issue time without a cutoff", &fakeRevoker{}, nil, http.StatusOK},
		{"redis error allows the request", &fakeRevoker{err: errors.New("redis down")}, at(cutoff), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, reached := serveAuthed(tt.revoker, signToken(t, testAccessSecret, tt.issuedAt))
			assert.Equal(t, tt.want, w.Code)
			assert.Equal(t, tt.want == http.StatusOK, reached, "handler reached")
			assert.Equal(t, "user-1", tt.revoker.userID, "checked user")
			assert.Equal(t, testSessionID, tt.revoker.sessionID, "checked session")
		})
	}
}

func TestRequireAuthSessionRevocation(t *testing.T) {
	now := jwt.NewNumericDate(time.Now())
	revoker := &fakeRevoker{revokedSession: "session-logged-out"}

	w, reached := serveAuthed(revoker, signSessionToken(t, testAccessSecret, "session-logged-out", now))
	assert.Equal(t, http.StatusUnauthorized, w.Code, "revoked session")
	assert.False(t, reached, "revoked session reached the handler")

	// The same user's other sessions keep working.
	w, reached = serveAuthed(revoker, signSessionToken(t, testAccessSecret, "session-other-device", now))
	assert.Equal(t, http.StatusOK, w.Code, "other session")
	assert.True(t, reached, "other session reached the handler")
}

func TestRequireAuthRejectsTokenWithoutSession(t *testing.T) {
	revoker := &fakeRevoker{}
	w, reached := serveAuthed(revoker, signSessionToken(t, testAccessSecret, "", jwt.NewNumericDate(time.Now())))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, reached, "handler reached")
	assert.Zero(t, revoker.calls, "a token without a session reached Redis")
}

func TestRequireAuthChecksSignatureBeforeRedis(t *testing.T) {
	revoker := &fakeRevoker{}
	w, _ := serveAuthed(revoker, signToken(t, "wrong-secret", jwt.NewNumericDate(time.Now())))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Zero(t, revoker.calls, "an invalid token reached Redis")
}
