package auth

import (
	"context"
	"time"
)

// TokenRevoker invalidates existing access tokens. Access tokens aren't recorded
// anywhere when issued, so this stores markers rather than a list of tokens: a
// cutoff time per user, and a flag per session. A marker only needs to last as long
// as an access token, so pass AccessTokenTTL as ttl.
type TokenRevoker interface {
	// RevokeTokensBefore rejects tokens the user was issued at or before cutoff, on
	// every session.
	RevokeTokensBefore(ctx context.Context, userID string, cutoff time.Time, ttl time.Duration) error
	// RevokeSession rejects every token issued for one session.
	RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error
	// Revocations returns the markers that apply to a token for this user and session.
	Revocations(ctx context.Context, userID, sessionID string) (Revocations, error)
}

// Revocations are the markers that apply to one access token.
type Revocations struct {
	// Cutoff is set when HasCutoff is true.
	Cutoff         time.Time
	HasCutoff      bool
	SessionRevoked bool
}

// Rejects reports whether these markers revoke the token.
func (r Revocations) Rejects(claims *AccessClaims) bool {
	return r.SessionRevoked || (r.HasCutoff && claims.IssuedAtOrBefore(r.Cutoff))
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
