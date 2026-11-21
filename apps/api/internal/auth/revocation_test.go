package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestIssuedAtOrBefore(t *testing.T) {
	cutoff := time.Unix(1_000, 500_000_000)

	tests := []struct {
		name     string
		issuedAt *jwt.NumericDate
		want     bool
	}{
		{"earlier", jwt.NewNumericDate(time.Unix(999, 0)), true},
		{"same second", jwt.NewNumericDate(time.Unix(1_000, 0)), true},
		{"later", jwt.NewNumericDate(time.Unix(1_001, 0)), false},
		{"no issue time", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := AccessClaims{RegisteredClaims: jwt.RegisteredClaims{IssuedAt: tt.issuedAt}}
			assert.Equal(t, tt.want, claims.IssuedAtOrBefore(cutoff))
		})
	}
}
