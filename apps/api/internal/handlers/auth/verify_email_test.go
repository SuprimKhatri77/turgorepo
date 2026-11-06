package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
)

func registerDeps(repo *fakeRepo, mailer *fakeMailer) AccountEmailDeps {
	return verifyDeps(repo, &fakeLimiter{}, mailer)
}

func verifyDeps(repo *fakeRepo, limiter *fakeLimiter, mailer *fakeMailer) AccountEmailDeps {
	cfg := testConfig()
	cfg.FrontendURL = "https://app.example.com/"
	cfg.EmailVerificationTTL = 24 * time.Hour
	return AccountEmailDeps{Queries: repo, Config: cfg, Limiter: limiter, Revoker: &fakeRevoker{}, Mailer: mailer, Jobs: inlineJobs{}}
}

func verifyTokenFromEmail(t *testing.T, msg email.Message) string {
	t.Helper()
	const marker = "https://app.example.com/auth/verify-email#token="
	_, rest, found := strings.Cut(msg.Text, marker)
	if !found {
		t.Fatalf("no verification link in email: %s", msg.Text)
	}
	return strings.Fields(rest)[0]
}

func TestVerifyEmail(t *testing.T) {
	body := func(token string) string { return `{"token":"` + token + `"}` }

	t.Run("verifies by the token's hash without logging in", func(t *testing.T) {
		repo := &fakeRepo{user: testUser(t)}
		w := serve(VerifyEmail(verifyDeps(repo, &fakeLimiter{}, &fakeMailer{})), body("the-token"))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body)
		}
		if repo.verifiedHash != session.HashToken("the-token") {
			t.Error("didn't look up the token by hash")
		}
		if len(setCookies(w)) != 0 || repo.sessionsNew != 0 {
			t.Error("verifying must not log the user in")
		}
	})

	t.Run("invalid, used, expired or outdated token", func(t *testing.T) {
		w := serve(VerifyEmail(verifyDeps(&fakeRepo{verifyErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{})), body("bad"))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", w.Code)
		}
		if code := decode(t, w).Code; code != constants.InvalidVerificationToken {
			t.Errorf("code = %q", code)
		}
	})

	t.Run("database error isn't reported as a bad link", func(t *testing.T) {
		w := serve(VerifyEmail(verifyDeps(&fakeRepo{verifyErr: errDB}, &fakeLimiter{}, &fakeMailer{})), body("x"))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("missing or oversized token is rejected before the database", func(t *testing.T) {
		for _, b := range []string{`{}`, body(strings.Repeat("a", 129))} {
			repo := &fakeRepo{}
			w := serve(VerifyEmail(verifyDeps(repo, &fakeLimiter{}, &fakeMailer{})), b)
			if w.Code != http.StatusBadRequest || repo.verifiedHash != "" {
				t.Errorf("%s: status = %d, looked up = %q", b, w.Code, repo.verifiedHash)
			}
		}
	})
}

func TestResendVerification(t *testing.T) {
	const body = `{"email":"Ada@Example.com"}`

	t.Run("sends a new link to an unverified account", func(t *testing.T) {
		repo := &fakeRepo{user: unverifiedUser(t)}
		mailer := &fakeMailer{}
		w := serve(ResendVerification(verifyDeps(repo, &fakeLimiter{}, mailer)), body)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if repo.lookedUp != "ada@example.com" {
			t.Errorf("looked up %q", repo.lookedUp)
		}
		if len(mailer.sent) != 1 {
			t.Fatalf("sent = %+v", mailer.sent)
		}
		if token := verifyTokenFromEmail(t, mailer.sent[0]); repo.verificationToken.TokenHash != session.HashToken(token) {
			t.Error("stored hash doesn't match the emailed token")
		}
	})

	t.Run("responds the same however it goes", func(t *testing.T) {
		tests := []struct {
			name      string
			repo      *fakeRepo
			limiter   *fakeLimiter
			mailer    *fakeMailer
			wantEmail bool
		}{
			{"unverified account", &fakeRepo{user: unverifiedUser(t)}, &fakeLimiter{}, &fakeMailer{}, true},
			{"already verified", &fakeRepo{user: testUser(t)}, &fakeLimiter{}, &fakeMailer{}, false},
			{"unknown email", &fakeRepo{getUserErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{}, false},
			{"database error", &fakeRepo{getUserErr: errDB}, &fakeLimiter{}, &fakeMailer{}, false},
			{"token not stored", &fakeRepo{user: unverifiedUser(t), verificationErr: errDB}, &fakeLimiter{}, &fakeMailer{}, false},
			{"email fails", &fakeRepo{user: unverifiedUser(t)}, &fakeLimiter{}, &fakeMailer{err: errors.New("smtp down")}, true},
			{"email limit reached", &fakeRepo{user: unverifiedUser(t)}, &fakeLimiter{blocked: true}, &fakeMailer{}, false},
		}

		var firstBody string
		for _, tt := range tests {
			w := serve(ResendVerification(verifyDeps(tt.repo, tt.limiter, tt.mailer)), body)
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
		serve(ResendVerification(verifyDeps(&fakeRepo{user: unverifiedUser(t)}, limiter, &fakeMailer{})), body)
		if len(limiter.keys) != 1 || !strings.HasPrefix(limiter.keys[0], "ratelimit:resend-verification:email:") {
			t.Fatalf("keys = %v", limiter.keys)
		}
		if strings.Contains(limiter.keys[0], "ada@example.com") {
			t.Errorf("key has the address: %s", limiter.keys[0])
		}
	})

	t.Run("invalid email is a validation error", func(t *testing.T) {
		w := serve(ResendVerification(verifyDeps(&fakeRepo{}, &fakeLimiter{}, &fakeMailer{})), `{"email":"nope"}`)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d", w.Code)
		}
	})
}
