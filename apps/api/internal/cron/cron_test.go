package cron

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakeStore struct {
	calls int
	err   error
}

func (s *fakeStore) DeleteExpiredSessions(context.Context) (pgconn.CommandTag, error) {
	s.calls++
	return pgconn.NewCommandTag("DELETE 2"), s.err
}

func TestCleanupExpiredSessions(t *testing.T) {
	store := &fakeStore{}
	cleanupExpiredSessions(store)
	if store.calls != 1 {
		t.Errorf("calls = %d", store.calls)
	}

	// A database error is logged rather than panicking inside the scheduler.
	cleanupExpiredSessions(&fakeStore{err: errors.New("down")})
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
