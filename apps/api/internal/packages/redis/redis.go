package redis

import (
	"context"
	"errors"
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
		if i+1 == maxRetries {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, fmt.Errorf("failed to connect to redis after %d attempts: %w", maxRetries, err)
}

// allowScript increments a fixed-window counter and returns its count and time
// left. It runs atomically so concurrent requests can't leave a key without an expiry.
var allowScript = goredis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl < 0 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
  ttl = tonumber(ARGV[1])
end
return {count, ttl}
`)

// Allow implements ratelimit.Limiter using a fixed window per key.
func (c *Client) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	values, err := allowScript.Run(ctx, c.rdb, []string{key}, window.Milliseconds()).Int64Slice()
	if err != nil {
		return false, 0, err
	}
	if len(values) != 2 {
		return false, 0, fmt.Errorf("rate limit script returned %d values", len(values))
	}
	return values[0] <= int64(limit), time.Duration(values[1]) * time.Millisecond, nil
}

// Clear implements ratelimit.Limiter by deleting a key's counter.
func (c *Client) Clear(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func tokenCutoffKey(userID string) string {
	return "auth:tokens_valid_after:" + userID
}

// RevokeTokensBefore implements auth.TokenRevoker.
func (c *Client) RevokeTokensBefore(ctx context.Context, userID string, cutoff time.Time, ttl time.Duration) error {
	return c.rdb.Set(ctx, tokenCutoffKey(userID), cutoff.Unix(), ttl).Err()
}

// TokensValidAfter implements auth.TokenRevoker.
func (c *Client) TokensValidAfter(ctx context.Context, userID string) (time.Time, bool, error) {
	seconds, err := c.rdb.Get(ctx, tokenCutoffKey(userID)).Int64()
	if errors.Is(err, goredis.Nil) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return time.Unix(seconds, 0), true, nil
}
