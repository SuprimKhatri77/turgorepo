package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

func TestForgotPassword(t *testing.T) {
	const body = `{"email":"Ada@Example.com"}`

	t.Run("emails a link and stores only the hash", func(t *testing.T) {
		repo := &resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}
		mailer := &fakeMailer{}
		w := serve(newHandler(repo, mailer).ForgotPassword, body)

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
			w := serve(newHandler(tt.repo, tt.limiter, tt.mailer).ForgotPassword, body)
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
		serve(newHandler(&resetRepo{fakeRepo: fakeRepo{user: testUser(t)}}, limiter).ForgotPassword, body)
		require.Len(t, limiter.keys, 1)
		require.True(t, strings.HasPrefix(limiter.keys[0], "ratelimit:forgot-password:email:"), "key = %s", limiter.keys[0])
		assert.NotContains(t, limiter.keys[0], "ada@example.com", "key has the address")
		assert.NotContains(t, limiter.keys[0], "192.0.2", "key has an IP")
	})

	t.Run("invalid email is a validation error", func(t *testing.T) {
		w := serve(newHandler(&resetRepo{}).ForgotPassword, `{"email":"nope"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
