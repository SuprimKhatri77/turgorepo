package cron

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	sessionCalls, tokenCalls, verificationCalls int
	sessionErr                                  error
}

func (s *fakeStore) DeleteExpiredSessions(context.Context) (pgconn.CommandTag, error) {
	s.sessionCalls++
	return pgconn.NewCommandTag("DELETE 2"), s.sessionErr
}

func (s *fakeStore) DeleteExpiredPasswordResetTokens(context.Context) (pgconn.CommandTag, error) {
	s.tokenCalls++
	return pgconn.NewCommandTag("DELETE 1"), nil
}

func (s *fakeStore) DeleteExpiredEmailVerificationTokens(context.Context) (pgconn.CommandTag, error) {
	s.verificationCalls++
	return pgconn.NewCommandTag("DELETE 3"), nil
}

func TestCleanupExpired(t *testing.T) {
	store := &fakeStore{}
	cleanupExpired(store)
	assert.Equal(t, 1, store.sessionCalls, "session cleanup calls")
	assert.Equal(t, 1, store.tokenCalls, "password reset token cleanup calls")
	assert.Equal(t, 1, store.verificationCalls, "email verification token cleanup calls")
}

func TestCleanupContinuesAfterAnError(t *testing.T) {
	store := &fakeStore{sessionErr: errors.New("down")}
	cleanupExpired(store)
	assert.Equal(t, 1, store.tokenCalls, "a failed job stopped the next ones")
	assert.Equal(t, 1, store.verificationCalls, "a failed job stopped the next ones")
}

func TestStartSchedulesJobs(t *testing.T) {
	c, err := Start(&fakeStore{})
	require.NoError(t, err)
	defer c.Stop()

	assert.Len(t, c.Entries(), 1)
}
