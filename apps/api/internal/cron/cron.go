// Package cron runs the API's scheduled jobs.
package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/robfig/cron/v3"
)

const cleanupTimeout = time.Minute

// SessionStore is the part of the database the jobs need.
type SessionStore interface {
	DeleteExpiredSessions(ctx context.Context) (pgconn.CommandTag, error)
}

// Start schedules the jobs and returns the running scheduler. Every API container
// runs its own scheduler, so jobs must be safe to run more than once.
func Start(store SessionStore) (*cron.Cron, error) {
	c := cron.New()

	if _, err := c.AddFunc("@daily", func() { cleanupExpiredSessions(store) }); err != nil {
		return nil, err
	}

	c.Start()
	return c, nil
}

func cleanupExpiredSessions(store SessionStore) {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	result, err := store.DeleteExpiredSessions(ctx)
	if err != nil {
		slog.Error("failed to delete expired sessions", "err", err)
		return
	}
	slog.Info("deleted expired sessions", "count", result.RowsAffected())
}
