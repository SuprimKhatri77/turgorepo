package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client wraps a go-redis client.
type Client struct {
	rdb *goredis.Client
}

// RDB exposes the underlying go-redis client for callers that need raw commands.
func (c *Client) RDB() *goredis.Client {
	return c.rdb
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// Ping checks connectivity.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func open(ctx context.Context, redisURL string) (*Client, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdb := goredis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

// ConnectWithRetry opens a Redis client and retries on failure.
func ConnectWithRetry(ctx context.Context, redisURL string, maxRetries int) (*Client, error) {
	var (
		client *Client
		err    error
	)
	for i := range maxRetries {
		client, err = open(ctx, redisURL)
		if err == nil {
			return client, nil
		}
		slog.Warn("failed to connect to redis, retrying...", "attempt", i+1, "max", maxRetries, "err", err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to redis after %d attempts: %w", maxRetries, err)
}
