package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	"golang.org/x/crypto/bcrypt"
)

// resetRepo adds the password reset queries to fakeRepo.
type resetRepo struct {
	fakeRepo

	createdToken db.CreatePasswordResetTokenParams
	createErr    error
	resetParams  db.ResetPasswordWithTokenParams
	resetErr     error
}

func (r *resetRepo) CreatePasswordResetToken(_ context.Context, params db.CreatePasswordResetTokenParams) error {
	r.createdToken = params
	return r.createErr
}

func (r *resetRepo) ResetPasswordWithToken(_ context.Context, params db.ResetPasswordWithTokenParams) (db.User, error) {
	r.resetParams = params
	return r.user, r.resetErr
}

type fakeMailer struct {
	sent []email.Message
	err  error
}

func (m *fakeMailer) Send(_ context.Context, msg email.Message) error {
	m.sent = append(m.sent, msg)
	return m.err
}

// inlineJobs runs background jobs immediately, so tests can check their effects.
type inlineJobs struct{}

func (inlineJobs) Go(_ string, job func(ctx context.Context)) { job(context.Background()) }

// fakeRevoker records the cutoffs and session revocations it's asked to set.
type fakeRevoker struct {
	userIDs  []string
	cutoff   time.Time
	ttl      time.Duration
	sessions []string
	err      error
}

func (r *fakeRevoker) RevokeTokensBefore(_ context.Context, userID string, cutoff time.Time, ttl time.Duration) error {
	r.userIDs = append(r.userIDs, userID)
	r.cutoff = cutoff
	r.ttl = ttl
	return r.err
}

func (r *fakeRevoker) RevokeSession(_ context.Context, sessionID string, ttl time.Duration) error {
	r.sessions = append(r.sessions, sessionID)
	r.ttl = ttl
	return r.err
}

func (r *fakeRevoker) Revocations(context.Context, string, string) (session.Revocations, error) {
	return session.Revocations{}, nil
}

func resetDeps(repo *resetRepo, limiter *fakeLimiter, mailer *fakeMailer) PasswordResetDeps {
	cfg := testConfig()
	cfg.FrontendURL = "https://app.example.com/"
	cfg.PasswordResetTTL = time.Hour
	return PasswordResetDeps{Queries: repo, Config: cfg, Limiter: limiter, Revoker: &fakeRevoker{}, Mailer: mailer, Jobs: inlineJobs{}}
}

func tokenFromEmail(t *testing.T, msg email.Message) string {
	t.Helper()
	const marker = "https://app.example.com/auth/reset-password#token="
	_, rest, found := strings.Cut(msg.Text, marker)
	if !found {
		t.Fatalf("no reset link in email: %s", msg.Text)
	}
	return strings.Fields(rest)[0]
}

func TestForgotPassword(t *testing.T) {
	const body = `{"email":"Ada@Example.com"}`

	t.Run("emails a link and stores only the hash", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		mailer := &fakeMailer{}
		w := serve(ForgotPassword(resetDeps(repo, &fakeLimiter{}, mailer)), body)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if repo.lookedUp != "ada@example.com" {
			t.Errorf("looked up %q", repo.lookedUp)
		}
		if len(mailer.sent) != 1 || mailer.sent[0].To != "ada@example.com" {
			t.Fatalf("sent = %+v", mailer.sent)
		}

		token := tokenFromEmail(t, mailer.sent[0])
		if repo.createdToken.TokenHash != session.HashToken(token) {
			t.Error("stored hash doesn't match the emailed token")
		}
		if repo.createdToken.TokenHash == token {
			t.Error("stored the raw token")
		}
		if remaining := time.Until(repo.createdToken.ExpiresAt.Time); remaining < 59*time.Minute || remaining > time.Hour {
			t.Errorf("expires in %v", remaining)
		}
	})

	t.Run("responds the same however it goes", func(t *testing.T) {
		tests := []struct {
			name      string
			repo      *resetRepo
			limiter   *fakeLimiter
			mailer    *fakeMailer
			wantEmail bool
		}{
			{"known email", &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}, &fakeLimiter{}, &fakeMailer{}, true},
			{"unknown email", &resetRepo{fakeRepo: fakeRepo{getUserErr: pgx.ErrNoRows}}, &fakeLimiter{}, &fakeMailer{}, false},
			{"database error", &resetRepo{fakeRepo: fakeRepo{getUserErr: errDB}}, &fakeLimiter{}, &fakeMailer{}, false},
			{"token not stored", &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}, createErr: errDB}, &fakeLimiter{}, &fakeMailer{}, false},
			{"email fails", &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}, &fakeLimiter{}, &fakeMailer{err: errors.New("smtp down")}, true},
			{"email limit reached", &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}, &fakeLimiter{blocked: true}, &fakeMailer{}, false},
		}

		var firstBody string
		for _, tt := range tests {
			w := serve(ForgotPassword(resetDeps(tt.repo, tt.limiter, tt.mailer)), body)
			if w.Code != http.StatusOK {
				t.Errorf("%s: status = %d", tt.name, w.Code)
			}
			if firstBody == "" {
				firstBody = w.Body.String()
			}
			if w.Body.String() != firstBody {
				t.Errorf("%s: body differs: %s", tt.name, w.Body)
			}
			if got := len(tt.mailer.sent) > 0; got != tt.wantEmail {
				t.Errorf("%s: emailed = %v, want %v", tt.name, got, tt.wantEmail)
			}
		}
	})

	t.Run("email limit applies across IPs, without the raw address", func(t *testing.T) {
		limiter := &fakeLimiter{}
		serve(ForgotPassword(resetDeps(&resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}, limiter, &fakeMailer{})), body)
		if len(limiter.keys) != 1 || !strings.HasPrefix(limiter.keys[0], "ratelimit:forgot-password:email:") {
			t.Fatalf("keys = %v", limiter.keys)
		}
		if strings.Contains(limiter.keys[0], "ada@example.com") || strings.Contains(limiter.keys[0], "192.0.2") {
			t.Errorf("key has the address or an IP: %s", limiter.keys[0])
		}
	})

	t.Run("invalid email is a validation error", func(t *testing.T) {
		w := serve(ForgotPassword(resetDeps(&resetRepo{}, &fakeLimiter{}, &fakeMailer{})), `{"email":"nope"}`)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d", w.Code)
		}
	})
}

func TestResetPassword(t *testing.T) {
	body := func(token, password string) string {
		return `{"token":"` + token + `","password":"` + password + `"}`
	}

	t.Run("sets the password, clears cookies and the login limit, and notifies", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		limiter := &fakeLimiter{}
		mailer := &fakeMailer{}
		w := serve(ResetPassword(resetDeps(repo, limiter, mailer)), body("the-token", "new-password-123"))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body)
		}
		if repo.resetParams.TokenHash != session.HashToken("the-token") {
			t.Error("didn't look up the token by hash")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(repo.resetParams.PasswordHash), []byte("new-password-123")); err != nil {
			t.Errorf("stored hash doesn't match the new password: %v", err)
		}
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			if value, set := setCookies(w)[name]; !set || value != "" {
				t.Errorf("cookie %s should be cleared, got %q (set=%v)", name, value, set)
			}
		}
		if len(limiter.cleared) != 1 || !strings.HasPrefix(limiter.cleared[0], "ratelimit:login:account:") {
			t.Errorf("cleared = %v", limiter.cleared)
		}
		if len(mailer.sent) != 1 || !strings.Contains(mailer.sent[0].Subject, "changed") {
			t.Errorf("sent = %+v", mailer.sent)
		}
	})

	t.Run("revokes the user's access tokens", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		deps := resetDeps(repo, &fakeLimiter{}, &fakeMailer{})
		revoker := &fakeRevoker{}
		deps.Revoker = revoker

		before := time.Now()
		serve(ResetPassword(deps), body("the-token", "new-password-123"))

		if len(revoker.userIDs) != 1 || revoker.userIDs[0] != "01010101-0101-0101-0101-010101010101" {
			t.Fatalf("userIDs = %v", revoker.userIDs)
		}
		if revoker.cutoff.Before(before) || revoker.cutoff.After(time.Now()) {
			t.Errorf("cutoff = %v", revoker.cutoff)
		}
		if revoker.ttl != deps.Config.AccessTokenTTL {
			t.Errorf("ttl = %v, want %v", revoker.ttl, deps.Config.AccessTokenTTL)
		}
	})

	t.Run("a failed revocation doesn't fail the reset", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		deps := resetDeps(repo, &fakeLimiter{}, &fakeMailer{})
		deps.Revoker = &fakeRevoker{err: errors.New("redis down")}
		if w := serve(ResetPassword(deps), body("the-token", "new-password-123")); w.Code != http.StatusOK {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("an invalid token revokes nothing", func(t *testing.T) {
		deps := resetDeps(&resetRepo{resetErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{})
		revoker := &fakeRevoker{}
		deps.Revoker = revoker
		serve(ResetPassword(deps), body("bad", "new-password-123"))
		if len(revoker.userIDs) != 0 {
			t.Errorf("userIDs = %v", revoker.userIDs)
		}
	})

	t.Run("a failed notification doesn't fail the reset", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		w := serve(ResetPassword(resetDeps(repo, &fakeLimiter{}, &fakeMailer{err: errors.New("smtp down")})), body("the-token", "new-password-123"))
		if w.Code != http.StatusOK {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("invalid or used token", func(t *testing.T) {
		mailer := &fakeMailer{}
		w := serve(ResetPassword(resetDeps(&resetRepo{resetErr: pgx.ErrNoRows}, &fakeLimiter{}, mailer)), body("bad", "new-password-123"))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", w.Code)
		}
		if code := decode(t, w).Code; code != constants.InvalidResetToken {
			t.Errorf("code = %q", code)
		}
		if len(setCookies(w)) != 0 {
			t.Error("a bad link must not log the user out")
		}
		if len(mailer.sent) != 0 {
			t.Error("sent an email for a failed reset")
		}
	})

	t.Run("database error isn't reported as a bad link", func(t *testing.T) {
		w := serve(ResetPassword(resetDeps(&resetRepo{resetErr: errDB}, &fakeLimiter{}, &fakeMailer{})), body("x", "new-password-123"))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("password rules match registration", func(t *testing.T) {
		tooLong := strings.Repeat("😀", 50)
		for _, password := range []string{"short", "", "        ", tooLong} {
			repo := &resetRepo{}
			w := serve(ResetPassword(resetDeps(repo, &fakeLimiter{}, &fakeMailer{})), body("the-token", password))
			if w.Code != http.StatusBadRequest {
				t.Errorf("%q: status = %d", password, w.Code)
			}
			if repo.resetParams.TokenHash != "" {
				t.Errorf("%q: token was used despite the invalid password", password)
			}
		}
	})

	t.Run("oversized token is rejected", func(t *testing.T) {
		w := serve(ResetPassword(resetDeps(&resetRepo{}, &fakeLimiter{}, &fakeMailer{})), body(strings.Repeat("a", 129), "new-password-123"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d", w.Code)
		}
	})
}

func TestResetTokensAreRandomAndURLSafe(t *testing.T) {
	first, firstHash, err := session.NewResetToken()
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := session.NewResetToken()
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Error("tokens repeat")
	}
	if len(first) != 43 {
		t.Errorf("length = %d, want 43 (32 bytes)", len(first))
	}
	if strings.ContainsAny(first, "+/=") {
		t.Errorf("token isn't URL safe: %s", first)
	}
	if firstHash != session.HashToken(first) {
		t.Error("hash doesn't match")
	}
}
