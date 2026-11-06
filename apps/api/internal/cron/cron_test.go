package cron

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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
	if store.sessionCalls != 1 || store.tokenCalls != 1 || store.verificationCalls != 1 {
		t.Errorf("calls = %+v", store)
	}
}

func TestCleanupContinuesAfterAnError(t *testing.T) {
	store := &fakeStore{sessionErr: errors.New("down")}
	cleanupExpired(store)
	if store.tokenCalls != 1 || store.verificationCalls != 1 {
		t.Error("a failed job stopped the next ones")
	}
}

func TestStartSchedulesJobs(t *testing.T) {
	c, err := Start(&fakeStore{})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Stop()

	if got := len(c.Entries()); got != 1 {
		t.Errorf("entries = %d", got)
	}
}
