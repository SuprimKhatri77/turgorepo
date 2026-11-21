package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	validator.Init()
	os.Exit(m.Run())
}

const testPassword = "password123"

var errDB = errors.New("connection refused")

// fakeRepo embeds the interface, so a call the test didn't expect panics.
type fakeRepo struct {
	repository.AuthRepository

	user        db.User
	getUserErr  error
	lookedUp    string
	created     *db.CreateUserParams
	createErr   error
	sessionErr  error
	sessions    []db.Session // returned in order by GetActiveSessionByID
	rotateErr   error
	rotated     bool
	revoked     []pgtype.UUID
	sessionsNew int
	newSessErr  error

	revokedUsers []pgtype.UUID
	revokeAllErr error

	verificationToken *db.CreateEmailVerificationTokenParams
	verificationErr   error
	verifiedHash      string
	verifyErr         error
}

func (r *fakeRepo) CreateEmailVerificationToken(_ context.Context, params db.CreateEmailVerificationTokenParams) error {
	r.verificationToken = &params
	return r.verificationErr
}

func (r *fakeRepo) VerifyEmailWithToken(_ context.Context, tokenHash string) (db.User, error) {
	r.verifiedHash = tokenHash
	return r.user, r.verifyErr
}

func (r *fakeRepo) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	r.lookedUp = email
	return r.user, r.getUserErr
}

func (r *fakeRepo) GetUserByID(context.Context, pgtype.UUID) (db.User, error) {
	return r.user, r.getUserErr
}

func (r *fakeRepo) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	r.created = &params
	if r.createErr != nil {
		return db.User{}, r.createErr
	}
	return db.User{ID: testUUID(1), Name: params.Name, Email: params.Email, Role: params.Role}, nil
}

func (r *fakeRepo) CreateSession(context.Context, db.CreateSessionParams) (db.Session, error) {
	r.sessionsNew++
	return db.Session{}, r.newSessErr
}

func (r *fakeRepo) GetActiveSessionByID(context.Context, pgtype.UUID) (db.Session, error) {
	if r.sessionErr != nil {
		return db.Session{}, r.sessionErr
	}
	if len(r.sessions) == 0 {
		return db.Session{}, pgx.ErrNoRows
	}
	s := r.sessions[0]
	r.sessions = r.sessions[1:]
	return s, nil
}

func (r *fakeRepo) RotateSessionToken(context.Context, db.RotateSessionTokenParams) (db.Session, error) {
	r.rotated = true
	return db.Session{}, r.rotateErr
}

func (r *fakeRepo) RevokeSession(_ context.Context, id pgtype.UUID) error {
	r.revoked = append(r.revoked, id)
	return nil
}

func (r *fakeRepo) RevokeAllUserSessions(_ context.Context, userID pgtype.UUID) error {
	r.revokedUsers = append(r.revokedUsers, userID)
	return r.revokeAllErr
}

type fakeLimiter struct {
	blocked bool
	keys    []string
	cleared []string
}

func (l *fakeLimiter) Clear(_ context.Context, key string) error {
	l.cleared = append(l.cleared, key)
	return nil
}

func (l *fakeLimiter) Allow(_ context.Context, key string, _ int, _ time.Duration) (bool, time.Duration, error) {
	l.keys = append(l.keys, key)
	return !l.blocked, time.Minute, nil
}

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

func TestLogin(t *testing.T) {
	t.Run("sets cookies and returns only public fields", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(Login(repo, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		cookies := setCookies(w)
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			assert.NotEmpty(t, cookies[name], "cookie %s not set", name)
		}

		var res struct {
			Data map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.Equal(t, "01010101-0101-0101-0101-010101010101", res.Data["id"])
		for _, key := range []string{"created_at", "updated_at", "image_url", "password_hash"} {
			assert.NotContains(t, res.Data, key)
		}
	})

	t.Run("refuses an unverified email without creating a session", func(t *testing.T) {
		repo := &fakeRepo{user: unverifiedUser(t)}
		w := serve(Login(repo, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)

		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.Equal(t, constants.EmailNotVerified, decode(t, w).Code)
		assert.Zero(t, repo.sessionsNew)
		assert.Empty(t, setCookies(w))
	})

	t.Run("an unverified account with the wrong password looks like any wrong password", func(t *testing.T) {
		unverified := serve(Login(&fakeRepo{user: unverifiedUser(t)}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"wrong-password"}`)
		verified := serve(Login(&fakeRepo{user: testUser(t)}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"wrong-password"}`)
		assert.Equal(t, http.StatusUnauthorized, unverified.Code)
		assert.Equal(t, verified.Body.String(), unverified.Body.String())
	})

	t.Run("looks up the normalized email", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		serve(Login(repo, testConfig(), &fakeLimiter{}), `{"email":"Ada@Example.COM","password":"`+testPassword+`"}`)
		assert.Equal(t, "ada@example.com", repo.lookedUp)
	})

	t.Run("wrong password and unknown email respond the same", func(t *testing.T) {
		wrong := serve(Login(&fakeRepo{user: testUser(t)}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"wrong-password"}`)
		unknown := serve(Login(&fakeRepo{getUserErr: pgx.ErrNoRows}, testConfig(), &fakeLimiter{}), `{"email":"no@example.com","password":"wrong-password"}`)

		require.Equal(t, http.StatusUnauthorized, wrong.Code, "wrong password")
		require.Equal(t, http.StatusUnauthorized, unknown.Code, "unknown email")
		assert.Equal(t, wrong.Body.String(), unknown.Body.String())
	})

	t.Run("database error is a 500, not invalid credentials", func(t *testing.T) {
		w := serve(Login(&fakeRepo{getUserErr: errDB}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("account limit blocks before checking the password", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(Login(repo, testConfig(), &fakeLimiter{blocked: true}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		require.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Empty(t, repo.lookedUp, "user was looked up despite the limit")
		assert.Equal(t, constants.RateLimited, decode(t, w).Code)
	})

	t.Run("account limit treats email case variants as one account", func(t *testing.T) {
		limiter := &fakeLimiter{}
		user := testUser(t)
		serve(Login(&fakeRepo{user: user}, testConfig(), limiter), `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		serve(Login(&fakeRepo{user: user}, testConfig(), limiter), `{"email":"ADA@example.com","password":"`+testPassword+`"}`)
		if assert.Len(t, limiter.keys, 2) {
			assert.Equal(t, limiter.keys[0], limiter.keys[1])
		}
		assert.NotContains(t, limiter.keys[0], "ada@example.com", "key contains the raw email")
	})

	t.Run("invalid body is rejected before the limiter and the lookup", func(t *testing.T) {
		tests := []struct {
			name      string
			body      string
			wantField string
		}{
			{"malformed JSON", `{"email":`, ""},
			{"missing email", `{"password":"` + testPassword + `"}`, "email"},
			{"invalid email", `{"email":"not-an-email","password":"` + testPassword + `"}`, "email"},
			{"missing password", `{"email":"ada@example.com"}`, "password"},
			{"password over 72 bytes", `{"email":"ada@example.com","password":"` + strings.Repeat("a", 73) + `"}`, "password"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := &fakeRepo{user: testUser(t)}
				limiter := &fakeLimiter{}
				w := serve(Login(repo, testConfig(), limiter), tt.body)

				require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				res := decode(t, w)
				assert.Equal(t, constants.ValidationFailed, res.Code)
				if tt.wantField != "" && assert.Len(t, res.Errors, 1) {
					assert.Equal(t, tt.wantField, res.Errors[0].Field)
				}
				assert.Empty(t, limiter.keys, "limiter was hit")
				assert.Empty(t, repo.lookedUp, "user was looked up")
			})
		}
	})

	t.Run("the password is not trimmed", func(t *testing.T) {
		w := serve(Login(&fakeRepo{user: testUser(t)}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":" `+testPassword+` "}`)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("a failed session insert is a 500 without cookies", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t), newSessErr: errDB}
		w := serve(Login(repo, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, constants.InternalServerError, decode(t, w).Code)
		assert.Empty(t, setCookies(w))
	})
}

func TestRegister(t *testing.T) {
	body := func(name, email, password string) string {
		b, err := json.Marshal(map[string]string{"name": name, "email": email, "password": password})
		require.NoError(t, err)
		return string(b)
	}

	t.Run("accepts real names", func(t *testing.T) {
		for _, name := range []string{"José García", "O'Brien", "O’Brien", "Mary-Jane", "सुप्रिम", "सुप्रिम खत्री", "J. R. Smith", "Martin Jr."} {
			w := serve(Register(registerDeps(&fakeRepo{}, &fakeMailer{})), body(name, "a@example.com", testPassword))
			assert.Equal(t, http.StatusCreated, w.Code, "%q: %s", name, w.Body)
		}
	})

	t.Run("rejects names that are not names", func(t *testing.T) {
		for _, name := range []string{"12345", "<script>", "John\nDoe", "--", "   "} {
			w := serve(Register(registerDeps(&fakeRepo{}, &fakeMailer{})), body(name, "a@example.com", testPassword))
			assert.Equal(t, http.StatusBadRequest, w.Code, "%q", name)
		}
	})

	t.Run("stores the normalized email", func(t *testing.T) {
		repo := &fakeRepo{}
		serve(Register(registerDeps(repo, &fakeMailer{})), body("Ada", "Ada.L@Example.COM", testPassword))
		if assert.NotNil(t, repo.created) {
			assert.Equal(t, "ada.l@example.com", repo.created.Email)
		}
	})

	t.Run("password over 72 bytes is a validation error", func(t *testing.T) {
		emoji := strings.Repeat("😀", 50) // 50 characters, 200 bytes
		w := serve(Register(registerDeps(&fakeRepo{}, &fakeMailer{})), body("Ada", "a@example.com", emoji))
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		errs := decode(t, w).Errors
		if assert.Len(t, errs, 1) {
			assert.Equal(t, "password", errs[0].Field)
		}
	})

	t.Run("password of exactly 72 bytes works", func(t *testing.T) {
		w := serve(Register(registerDeps(&fakeRepo{}, &fakeMailer{})), body("Ada", "a@example.com", strings.Repeat("a", 72)))
		assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	})

	t.Run("duplicate email is a conflict", func(t *testing.T) {
		repo := &fakeRepo{createErr: &pgconn.PgError{Code: "23505"}}
		w := serve(Register(registerDeps(repo, &fakeMailer{})), body("Ada", "a@example.com", testPassword))
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("emails a verification link and doesn't log the user in", func(t *testing.T) {
		repo := &fakeRepo{}
		mailer := &fakeMailer{}
		w := serve(Register(registerDeps(repo, mailer)), body("Ada", "Ada@Example.com", testPassword))

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		assert.Empty(t, setCookies(w))
		assert.Zero(t, repo.sessionsNew)
		assert.Nil(t, decode(t, w).Data, "response should not include the user")
		require.Len(t, mailer.sent, 1)
		require.Equal(t, "ada@example.com", mailer.sent[0].To)
		token := verifyTokenFromEmail(t, mailer.sent[0])
		stored := repo.verificationToken
		require.NotNil(t, stored)
		require.Equal(t, session.HashToken(token), stored.TokenHash)
		require.Equal(t, "ada@example.com", stored.Email)
		assert.WithinRange(t, stored.ExpiresAt.Time, time.Now().Add(23*time.Hour), time.Now().Add(24*time.Hour))
	})

	t.Run("a failed email doesn't fail registration", func(t *testing.T) {
		w := serve(Register(registerDeps(&fakeRepo{}, &fakeMailer{err: errors.New("smtp down")})), body("Ada", "a@example.com", testPassword))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("new users are always members", func(t *testing.T) {
		repo := &fakeRepo{}
		serve(Register(registerDeps(repo, &fakeMailer{})), `{"name":"Ada","email":"a@example.com","password":"`+testPassword+`","role":"admin"}`)
		if assert.NotNil(t, repo.created) {
			assert.Equal(t, constants.RoleMember, repo.created.Role)
		}
	})

	t.Run("invalid fields are rejected and nothing is created", func(t *testing.T) {
		tests := []struct {
			name, body, wantField string
		}{
			{"missing name", `{"email":"a@example.com","password":"` + testPassword + `"}`, "name"},
			{"name too short", body("A", "a@example.com", testPassword), "name"},
			{"name too long", body(strings.Repeat("a", 51), "a@example.com", testPassword), "name"},
			{"missing email", `{"name":"Ada","password":"` + testPassword + `"}`, "email"},
			{"invalid email", body("Ada", "not-an-email", testPassword), "email"},
			{"missing password", `{"name":"Ada","email":"a@example.com"}`, "password"},
			{"password under 8 characters", body("Ada", "a@example.com", "short"), "password"},
			{"blank password", body("Ada", "a@example.com", strings.Repeat(" ", 8)), "password"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := &fakeRepo{}
				mailer := &fakeMailer{}
				w := serve(Register(registerDeps(repo, mailer)), tt.body)

				require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				res := decode(t, w)
				assert.Equal(t, constants.ValidationFailed, res.Code)
				if assert.Len(t, res.Errors, 1) {
					assert.Equal(t, tt.wantField, res.Errors[0].Field)
				}
				assert.Nil(t, repo.created, "user was created")
				assert.Empty(t, mailer.sent)
			})
		}
	})

	t.Run("malformed JSON is a validation error", func(t *testing.T) {
		repo := &fakeRepo{}
		w := serve(Register(registerDeps(repo, &fakeMailer{})), `{"name":`)
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, constants.ValidationFailed, decode(t, w).Code)
		assert.Nil(t, repo.created)
	})

	t.Run("stores a bcrypt hash of the untrimmed password", func(t *testing.T) {
		repo := &fakeRepo{}
		password := "  spaces around  "
		serve(Register(registerDeps(repo, &fakeMailer{})), body("Ada", "a@example.com", password))

		require.NotNil(t, repo.created)
		assert.NotEqual(t, password, repo.created.PasswordHash)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.created.PasswordHash), []byte(password)))
	})

	t.Run("other database errors are a 500 and send no email", func(t *testing.T) {
		for _, err := range []error{errDB, &pgconn.PgError{Code: "23503"}} {
			mailer := &fakeMailer{}
			w := serve(Register(registerDeps(&fakeRepo{createErr: err}, mailer)), body("Ada", "a@example.com", testPassword))
			assert.Equal(t, http.StatusInternalServerError, w.Code, "%v", err)
			assert.Equal(t, constants.InternalServerError, decode(t, w).Code, "%v", err)
			assert.Empty(t, mailer.sent, "%v", err)
		}
	})
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
