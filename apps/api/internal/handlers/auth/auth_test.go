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
	return db.Session{}, nil
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
	if err != nil {
		t.Fatal(err)
	}
	return db.User{
		ID:           testUUID(1),
		Name:         "Ada Lovelace",
		Email:        "ada@example.com",
		PasswordHash: string(hash),
		Role:         constants.RoleMember,
	}
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
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode %q: %v", w.Body.String(), err)
	}
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

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body)
		}
		cookies := setCookies(w)
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			if cookies[name] == "" {
				t.Errorf("cookie %s not set", name)
			}
		}

		var res struct {
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if res.Data["id"] != "01010101-0101-0101-0101-010101010101" {
			t.Errorf("id = %v", res.Data["id"])
		}
		for _, key := range []string{"created_at", "updated_at", "image_url", "password_hash"} {
			if _, present := res.Data[key]; present {
				t.Errorf("response should not contain %s", key)
			}
		}
	})

	t.Run("looks up the normalized email", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		serve(Login(repo, testConfig(), &fakeLimiter{}), `{"email":"Ada@Example.COM","password":"`+testPassword+`"}`)
		if repo.lookedUp != "ada@example.com" {
			t.Errorf("looked up %q", repo.lookedUp)
		}
	})

	t.Run("wrong password and unknown email respond the same", func(t *testing.T) {
		wrong := serve(Login(&fakeRepo{user: testUser(t)}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"wrong-password"}`)
		unknown := serve(Login(&fakeRepo{getUserErr: pgx.ErrNoRows}, testConfig(), &fakeLimiter{}), `{"email":"no@example.com","password":"wrong-password"}`)

		if wrong.Code != http.StatusUnauthorized || unknown.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d and %d", wrong.Code, unknown.Code)
		}
		if wrong.Body.String() != unknown.Body.String() {
			t.Errorf("bodies differ: %s vs %s", wrong.Body, unknown.Body)
		}
	})

	t.Run("database error is a 500, not invalid credentials", func(t *testing.T) {
		w := serve(Login(&fakeRepo{getUserErr: errDB}, testConfig(), &fakeLimiter{}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("account limit blocks before checking the password", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(Login(repo, testConfig(), &fakeLimiter{blocked: true}), `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d", w.Code)
		}
		if repo.lookedUp != "" {
			t.Error("user was looked up despite the limit")
		}
		if code := decode(t, w).Code; code != constants.RateLimited {
			t.Errorf("code = %q", code)
		}
	})

	t.Run("account limit treats email case variants as one account", func(t *testing.T) {
		limiter := &fakeLimiter{}
		user := testUser(t)
		serve(Login(&fakeRepo{user: user}, testConfig(), limiter), `{"email":"ada@example.com","password":"`+testPassword+`"}`)
		serve(Login(&fakeRepo{user: user}, testConfig(), limiter), `{"email":"ADA@example.com","password":"`+testPassword+`"}`)
		if len(limiter.keys) != 2 || limiter.keys[0] != limiter.keys[1] {
			t.Errorf("keys = %v", limiter.keys)
		}
		if strings.Contains(limiter.keys[0], "ada@example.com") {
			t.Error("key contains the raw email")
		}
	})
}

func TestRegister(t *testing.T) {
	body := func(name, email, password string) string {
		b, err := json.Marshal(map[string]string{"name": name, "email": email, "password": password})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	t.Run("accepts real names", func(t *testing.T) {
		for _, name := range []string{"José García", "O'Brien", "O’Brien", "Mary-Jane", "सुप्रिम", "सुप्रिम खत्री", "J. R. Smith", "Martin Jr."} {
			w := serve(Register(&fakeRepo{}, testConfig()), body(name, "a@example.com", testPassword))
			if w.Code != http.StatusCreated {
				t.Errorf("%q: status = %d: %s", name, w.Code, w.Body)
			}
		}
	})

	t.Run("rejects names that are not names", func(t *testing.T) {
		for _, name := range []string{"12345", "<script>", "John\nDoe", "--", "   "} {
			w := serve(Register(&fakeRepo{}, testConfig()), body(name, "a@example.com", testPassword))
			if w.Code != http.StatusBadRequest {
				t.Errorf("%q: status = %d", name, w.Code)
			}
		}
	})

	t.Run("stores the normalized email", func(t *testing.T) {
		repo := &fakeRepo{}
		serve(Register(repo, testConfig()), body("Ada", "Ada.L@Example.COM", testPassword))
		if repo.created == nil || repo.created.Email != "ada.l@example.com" {
			t.Errorf("created = %+v", repo.created)
		}
	})

	t.Run("password over 72 bytes is a validation error", func(t *testing.T) {
		emoji := strings.Repeat("😀", 50) // 50 characters, 200 bytes
		w := serve(Register(&fakeRepo{}, testConfig()), body("Ada", "a@example.com", emoji))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d: %s", w.Code, w.Body)
		}
		errs := decode(t, w).Errors
		if len(errs) != 1 || errs[0].Field != "password" {
			t.Errorf("errors = %+v", errs)
		}
	})

	t.Run("password of exactly 72 bytes works", func(t *testing.T) {
		w := serve(Register(&fakeRepo{}, testConfig()), body("Ada", "a@example.com", strings.Repeat("a", 72)))
		if w.Code != http.StatusCreated {
			t.Errorf("status = %d: %s", w.Code, w.Body)
		}
	})

	t.Run("duplicate email is a conflict", func(t *testing.T) {
		repo := &fakeRepo{createErr: &pgconn.PgError{Code: "23505"}}
		w := serve(Register(repo, testConfig()), body("Ada", "a@example.com", testPassword))
		if w.Code != http.StatusConflict {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("new users are always members", func(t *testing.T) {
		repo := &fakeRepo{}
		serve(Register(repo, testConfig()), `{"name":"Ada","email":"a@example.com","password":"`+testPassword+`","role":"admin"}`)
		if repo.created == nil || repo.created.Role != constants.RoleMember {
			t.Errorf("created = %+v", repo.created)
		}
	})
}

// refreshFixture is a session whose current token is the returned token.
func refreshFixture(t *testing.T) (db.User, string, db.Session) {
	t.Helper()
	cfg := testConfig()
	user := testUser(t)
	tokens, err := session.NewTokens(cfg, user, testUUID(2))
	if err != nil {
		t.Fatal(err)
	}
	return user, tokens.RefreshToken, db.Session{
		ID:               testUUID(2),
		UserID:           user.ID,
		CurrentTokenHash: session.HashToken(tokens.RefreshToken),
	}
}

func refreshCookie(token string) *http.Cookie {
	return &http.Cookie{Name: "refresh_token", Value: token}
}

func TestRefresh(t *testing.T) {
	t.Run("current token rotates", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		w := serve(Refresh(repo, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusOK || !repo.rotated {
			t.Fatalf("status = %d, rotated = %v", w.Code, repo.rotated)
		}
		cookies := setCookies(w)
		if cookies["access_token"] == "" || cookies["refresh_token"] == "" {
			t.Errorf("cookies = %v", cookies)
		}
	})

	t.Run("previous token in the grace window only gets an access token", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		sess.PreviousTokenHash = pgtype.Text{String: sess.CurrentTokenHash, Valid: true}
		sess.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}
		sess.CurrentTokenHash = "hash-of-newer-token"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		w := serve(Refresh(repo, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusOK || repo.rotated {
			t.Fatalf("status = %d, rotated = %v", w.Code, repo.rotated)
		}
		cookies := setCookies(w)
		if cookies["access_token"] == "" {
			t.Error("access token not set")
		}
		if _, set := cookies["refresh_token"]; set {
			t.Error("refresh token should not change")
		}
	})

	t.Run("previous token after the grace window revokes the session", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		sess.PreviousTokenHash = pgtype.Text{String: sess.CurrentTokenHash, Valid: true}
		sess.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true}
		sess.CurrentTokenHash = "hash-of-newer-token"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess}}
		w := serve(Refresh(repo, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusUnauthorized || len(repo.revoked) != 1 {
			t.Errorf("status = %d, revoked = %v", w.Code, repo.revoked)
		}
	})

	t.Run("expired or revoked session clears cookies", func(t *testing.T) {
		user, token, _ := refreshFixture(t)
		w := serve(Refresh(&fakeRepo{user: user, sessionErr: pgx.ErrNoRows}, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", w.Code)
		}
		if value, set := setCookies(w)["refresh_token"]; !set || value != "" {
			t.Errorf("refresh cookie should be cleared, got %q (set=%v)", value, set)
		}
	})

	t.Run("database outage keeps the user logged in", func(t *testing.T) {
		user, token, _ := refreshFixture(t)
		w := serve(Refresh(&fakeRepo{user: user, sessionErr: errDB}, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d", w.Code)
		}
		if cookies := setCookies(w); len(cookies) != 0 {
			t.Errorf("cookies should be untouched, got %v", cookies)
		}
		if code := decode(t, w).Code; code != constants.ServiceUnavailable {
			t.Errorf("code = %q", code)
		}
	})

	t.Run("losing the rotation race falls back to the grace path", func(t *testing.T) {
		user, token, sess := refreshFixture(t)
		reloaded := sess
		reloaded.PreviousTokenHash = pgtype.Text{String: sess.CurrentTokenHash, Valid: true}
		reloaded.PreviousRotatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		reloaded.CurrentTokenHash = "hash-of-winner"
		repo := &fakeRepo{user: user, sessions: []db.Session{sess, reloaded}, rotateErr: pgx.ErrNoRows}
		w := serve(Refresh(repo, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body)
		}
		if _, set := setCookies(w)["refresh_token"]; set {
			t.Error("the losing request must not set a refresh token")
		}
	})

	t.Run("missing and invalid cookies are rejected", func(t *testing.T) {
		if w := serve(Refresh(&fakeRepo{}, testConfig()), ""); w.Code != http.StatusUnauthorized {
			t.Errorf("missing: status = %d", w.Code)
		}
		if w := serve(Refresh(&fakeRepo{}, testConfig()), "", refreshCookie("not-a-jwt")); w.Code != http.StatusUnauthorized {
			t.Errorf("invalid: status = %d", w.Code)
		}
	})
}

func TestLogout(t *testing.T) {
	t.Run("revokes the session and clears cookies", func(t *testing.T) {
		_, token, _ := refreshFixture(t)
		repo := &fakeRepo{}
		w := serve(Logout(repo, testConfig()), "", refreshCookie(token))

		if w.Code != http.StatusOK || len(repo.revoked) != 1 || repo.revoked[0] != testUUID(2) {
			t.Fatalf("status = %d, revoked = %v", w.Code, repo.revoked)
		}
		if value, set := setCookies(w)["access_token"]; !set || value != "" {
			t.Errorf("access cookie should be cleared, got %q (set=%v)", value, set)
		}
	})

	t.Run("works without a session", func(t *testing.T) {
		repo := &fakeRepo{}
		for _, cookies := range [][]*http.Cookie{nil, {refreshCookie("not-a-jwt")}} {
			if w := serve(Logout(repo, testConfig()), "", cookies...); w.Code != http.StatusOK {
				t.Errorf("status = %d", w.Code)
			}
		}
		if len(repo.revoked) != 0 {
			t.Errorf("revoked = %v", repo.revoked)
		}
	})
}
