package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func new(ctx context.Context, connString string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	config.MaxConnIdleTime = 4 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

// Close closes the connection pool.
func (db *DB) Close() {
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}
}

func ConnectWithRetry(ctx context.Context, connString string, maxRetries int) (*DB, error) {
	var (
		database *DB
		err      error
	)
	for i := range maxRetries {
		database, err = new(ctx, connString)
		if err == nil {
			return database, nil
		}
		slog.Warn("failed to connect to database, retrying...", "attempt", i+1, "max", maxRetries, "err", err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to open database after %d attempts: %w", maxRetries, err)
}
