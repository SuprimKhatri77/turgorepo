package utils

import "strings"

// NormalizeEmail returns the canonical form stored in users.email, so addresses
// differing only by case resolve to one account.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
