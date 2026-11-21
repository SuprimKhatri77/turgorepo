package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func resetDeps(repo *resetRepo, limiter *fakeLimiter, mailer *fakeMailer) AccountEmailDeps {
	cfg := testConfig()
	cfg.FrontendURL = "https://app.example.com/"
	cfg.PasswordResetTTL = time.Hour
	return AccountEmailDeps{Queries: repo, Config: cfg, Limiter: limiter, Revoker: &fakeRevoker{}, Mailer: mailer, Jobs: inlineJobs{}}
}

func tokenFromEmail(t *testing.T, msg email.Message) string {
	t.Helper()
	const marker = "https://app.example.com/auth/reset-password#token="
	_, rest, found := strings.Cut(msg.Text, marker)
	require.True(t, found, "no reset link in email: %s", msg.Text)
	return strings.Fields(rest)[0]
}

func TestForgotPassword(t *testing.T) {
	const body = `{"email":"Ada@Example.com"}`

	t.Run("emails a link and stores only the hash", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		mailer := &fakeMailer{}
		w := serve(ForgotPassword(resetDeps(repo, &fakeLimiter{}, mailer)), body)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ada@example.com", repo.lookedUp)
		require.Len(t, mailer.sent, 1)
		require.Equal(t, "ada@example.com", mailer.sent[0].To)

		token := tokenFromEmail(t, mailer.sent[0])
		assert.Equal(t, session.HashToken(token), repo.createdToken.TokenHash, "stored hash doesn't match the emailed token")
		assert.NotEqual(t, token, repo.createdToken.TokenHash, "stored the raw token")
		assert.WithinRange(t, repo.createdToken.ExpiresAt.Time, time.Now().Add(59*time.Minute), time.Now().Add(time.Hour))
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
			assert.Equal(t, http.StatusOK, w.Code, tt.name)
			if firstBody == "" {
				firstBody = w.Body.String()
			}
			assert.Equal(t, firstBody, w.Body.String(), "%s: body differs", tt.name)
			assert.Equal(t, tt.wantEmail, len(tt.mailer.sent) > 0, "%s: emailed", tt.name)
		}
	})

	t.Run("email limit applies across IPs, without the raw address", func(t *testing.T) {
		limiter := &fakeLimiter{}
		serve(ForgotPassword(resetDeps(&resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}, limiter, &fakeMailer{})), body)
		require.Len(t, limiter.keys, 1)
		require.True(t, strings.HasPrefix(limiter.keys[0], "ratelimit:forgot-password:email:"), "key = %s", limiter.keys[0])
		assert.NotContains(t, limiter.keys[0], "ada@example.com", "key has the address")
		assert.NotContains(t, limiter.keys[0], "192.0.2", "key has an IP")
	})

	t.Run("invalid email is a validation error", func(t *testing.T) {
		w := serve(ForgotPassword(resetDeps(&resetRepo{}, &fakeLimiter{}, &fakeMailer{})), `{"email":"nope"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
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

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, session.HashToken("the-token"), repo.resetParams.TokenHash, "didn't look up the token by hash")
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.resetParams.PasswordHash), []byte("new-password-123")), "stored hash doesn't match the new password")
		for _, name := range []string{"access_token", "refresh_token", "is_logged_in"} {
			assert.Contains(t, setCookies(w), name, "cookie %s should be cleared", name)
			assert.Empty(t, setCookies(w)[name], "cookie %s should be cleared", name)
		}
		if assert.Len(t, limiter.cleared, 1) {
			assert.True(t, strings.HasPrefix(limiter.cleared[0], "ratelimit:login:account:"), "cleared = %v", limiter.cleared)
		}
		if assert.Len(t, mailer.sent, 1) {
			assert.Contains(t, mailer.sent[0].Subject, "changed")
		}
	})

	t.Run("revokes the user's access tokens", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		deps := resetDeps(repo, &fakeLimiter{}, &fakeMailer{})
		revoker := &fakeRevoker{}
		deps.Revoker = revoker

		before := time.Now()
		serve(ResetPassword(deps), body("the-token", "new-password-123"))

		require.Equal(t, []string{"01010101-0101-0101-0101-010101010101"}, revoker.userIDs)
		assert.WithinRange(t, revoker.cutoff, before, time.Now())
		assert.Equal(t, deps.Config.AccessTokenTTL, revoker.ttl)
	})

	t.Run("a failed revocation doesn't fail the reset", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		deps := resetDeps(repo, &fakeLimiter{}, &fakeMailer{})
		deps.Revoker = &fakeRevoker{err: errors.New("redis down")}
		assert.Equal(t, http.StatusOK, serve(ResetPassword(deps), body("the-token", "new-password-123")).Code)
	})

	t.Run("an invalid token revokes nothing", func(t *testing.T) {
		deps := resetDeps(&resetRepo{resetErr: pgx.ErrNoRows}, &fakeLimiter{}, &fakeMailer{})
		revoker := &fakeRevoker{}
		deps.Revoker = revoker
		serve(ResetPassword(deps), body("bad", "new-password-123"))
		assert.Empty(t, revoker.userIDs)
	})

	t.Run("a failed notification doesn't fail the reset", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		w := serve(ResetPassword(resetDeps(repo, &fakeLimiter{}, &fakeMailer{err: errors.New("smtp down")})), body("the-token", "new-password-123"))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid or used token", func(t *testing.T) {
		mailer := &fakeMailer{}
		w := serve(ResetPassword(resetDeps(&resetRepo{resetErr: pgx.ErrNoRows}, &fakeLimiter{}, mailer)), body("bad", "new-password-123"))

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, constants.InvalidResetToken, decode(t, w).Code)
		assert.Empty(t, setCookies(w), "a bad link must not log the user out")
		assert.Empty(t, mailer.sent, "sent an email for a failed reset")
	})

	t.Run("database error isn't reported as a bad link", func(t *testing.T) {
		w := serve(ResetPassword(resetDeps(&resetRepo{resetErr: errDB}, &fakeLimiter{}, &fakeMailer{})), body("x", "new-password-123"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("password rules match registration", func(t *testing.T) {
		tooLong := strings.Repeat("😀", 50)
		for _, password := range []string{"short", "", "        ", tooLong} {
			repo := &resetRepo{}
			w := serve(ResetPassword(resetDeps(repo, &fakeLimiter{}, &fakeMailer{})), body("the-token", password))
			assert.Equal(t, http.StatusBadRequest, w.Code, "%q", password)
			assert.Empty(t, repo.resetParams.TokenHash, "%q: token was used despite the invalid password", password)
		}
	})

	t.Run("oversized token is rejected", func(t *testing.T) {
		w := serve(ResetPassword(resetDeps(&resetRepo{}, &fakeLimiter{}, &fakeMailer{})), body(strings.Repeat("a", 129), "new-password-123"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEmailTokensAreRandomAndURLSafe(t *testing.T) {
	first, firstHash, err := session.NewEmailToken()
	require.NoError(t, err)
	second, _, err := session.NewEmailToken()
	require.NoError(t, err)

	assert.NotEqual(t, first, second, "tokens repeat")
	assert.Len(t, first, 43, "32 bytes")
	assert.False(t, strings.ContainsAny(first, "+/="), "token isn't URL safe: %s", first)
	assert.Equal(t, session.HashToken(first), firstHash, "hash doesn't match")
}
