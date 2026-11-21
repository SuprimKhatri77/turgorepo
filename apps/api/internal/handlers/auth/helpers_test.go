package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"golang.org/x/crypto/bcrypt"
)

const testPassword = "password123"

var errDB = errors.New("connection refused")

func testUUID(b byte) pgtype.UUID {
	var id uuid.UUID
	for i := range id {
		id[i] = b
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func testConfig() *config.Config {
	return &config.Config{
		JWTAccessSecret:    "access-secret",
		JWTRefreshSecret:   "refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    time.Hour,
		RefreshReuseWindow: 5 * time.Minute,
	}
}

func testUser(t *testing.T) db.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	require.NoError(t, err)
	return db.User{
		ID:              testUUID(1),
		Name:            "Ada Lovelace",
		Email:           "ada@example.com",
		PasswordHash:    string(hash),
		Role:            constants.RoleMember,
		EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
	}
}

func unverifiedUser(t *testing.T) db.User {
	t.Helper()
	user := testUser(t)
	user.EmailVerifiedAt = pgtype.Timestamptz{}
	return user
}

func serve(handler gin.HandlerFunc, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := gin.New()
	r.POST("/", handler)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) types.APIResponse {
	t.Helper()
	var res types.APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res), "decode %q", w.Body.String())
	return res
}

// setCookies maps cookie name to value for cookies set on the response.
func setCookies(w *httptest.ResponseRecorder) map[string]string {
	found := map[string]string{}
	for _, c := range w.Result().Cookies() {
		found[c.Name] = c.Value
	}
	return found
}

func registerDeps(repo *fakeRepo, mailer *fakeMailer) AccountEmailDeps {
	return verifyDeps(repo, &fakeLimiter{}, mailer)
}

func verifyDeps(repo *fakeRepo, limiter *fakeLimiter, mailer *fakeMailer) AccountEmailDeps {
	cfg := testConfig()
	cfg.FrontendURL = "https://app.example.com/"
	cfg.EmailVerificationTTL = 24 * time.Hour
	return AccountEmailDeps{Queries: repo, Config: cfg, Limiter: limiter, Revoker: &fakeRevoker{}, Mailer: mailer, Jobs: inlineJobs{}}
}

func resetDeps(repo *resetRepo, limiter *fakeLimiter, mailer *fakeMailer) AccountEmailDeps {
	cfg := testConfig()
	cfg.FrontendURL = "https://app.example.com/"
	cfg.PasswordResetTTL = time.Hour
	return AccountEmailDeps{Queries: repo, Config: cfg, Limiter: limiter, Revoker: &fakeRevoker{}, Mailer: mailer, Jobs: inlineJobs{}}
}

func verifyTokenFromEmail(t *testing.T, msg email.Message) string {
	t.Helper()
	const marker = "https://app.example.com/auth/verify-email#token="
	_, rest, found := strings.Cut(msg.Text, marker)
	require.True(t, found, "no verification link in email: %s", msg.Text)
	return strings.Fields(rest)[0]
}

func tokenFromEmail(t *testing.T, msg email.Message) string {
	t.Helper()
	const marker = "https://app.example.com/auth/reset-password#token="
	_, rest, found := strings.Cut(msg.Text, marker)
	require.True(t, found, "no reset link in email: %s", msg.Text)
	return strings.Fields(rest)[0]
}

// refreshFixture is a session whose current token is the returned token.
func refreshFixture(t *testing.T) (db.User, string, db.Session) {
	t.Helper()
	cfg := testConfig()
	user := testUser(t)
	tokens, err := session.NewTokens(cfg, user, testUUID(2))
	require.NoError(t, err)
	return user, tokens.RefreshToken, db.Session{
		ID:               testUUID(2),
		UserID:           user.ID,
		CurrentTokenHash: session.HashToken(tokens.RefreshToken),
	}
}

// testSessionID is the string form of testUUID(2), the session refreshFixture uses.
const testSessionID = "02020202-0202-0202-0202-020202020202"

func refreshCookie(token string) *http.Cookie {
	return &http.Cookie{Name: "refresh_token", Value: token}
}
