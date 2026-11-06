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

// Store is the part of the database the jobs need.
type Store interface {
	DeleteExpiredSessions(ctx context.Context) (pgconn.CommandTag, error)
	DeleteExpiredPasswordResetTokens(ctx context.Context) (pgconn.CommandTag, error)
	DeleteExpiredEmailVerificationTokens(ctx context.Context) (pgconn.CommandTag, error)
}

// Start schedules the jobs and returns the running scheduler. Every API container
// runs its own scheduler, so jobs must be safe to run more than once.
func Start(store Store) (*cron.Cron, error) {
	c := cron.New()

	if _, err := c.AddFunc("@daily", func() { cleanupExpired(store) }); err != nil {
		return nil, err
	}

	c.Start()
	return c, nil
}

func cleanupExpired(store Store) {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	jobs := []struct {
		name string
		run  func(context.Context) (pgconn.CommandTag, error)
	}{
		{"sessions", store.DeleteExpiredSessions},
		{"password reset tokens", store.DeleteExpiredPasswordResetTokens},
		{"email verification tokens", store.DeleteExpiredEmailVerificationTokens},
	}

	// One failing doesn't stop the others.
	for _, job := range jobs {
		result, err := job.run(ctx)
		if err != nil {
			slog.Error("failed to delete expired rows", "table", job.name, "err", err)
			continue
		}
		slog.Info("deleted expired rows", "table", job.name, "count", result.RowsAffected())
	}
}
