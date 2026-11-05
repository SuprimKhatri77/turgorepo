package auth

import (
	"context"
	"time"
)

// TokenRevoker invalidates a user's existing access tokens. Access tokens aren't
// recorded anywhere when issued, so this stores a cutoff time for the user rather
// than a list of tokens. Each cutoff only needs to last as long as an access token.
type TokenRevoker interface {
	// RevokeTokensBefore rejects tokens the user was issued at or before cutoff.
	RevokeTokensBefore(ctx context.Context, userID string, cutoff time.Time, ttl time.Duration) error
	// TokensValidAfter returns the user's cutoff, and whether one is set.
	TokensValidAfter(ctx context.Context, userID string) (time.Time, bool, error)
}

// IssuedAtOrBefore reports whether the token was issued at or before cutoff. JWT
// times have one-second precision, so a token from the same second counts as older.
// A token with no issue time also counts as older.
func (c *AccessClaims) IssuedAtOrBefore(cutoff time.Time) bool {
	if c.IssuedAt == nil {
		return true
	}
	return c.IssuedAt.Unix() <= cutoff.Unix()
}
