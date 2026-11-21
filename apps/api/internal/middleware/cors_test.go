package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

func TestAllowedOrigins(t *testing.T) {
	got := allowedOrigins(&config.Config{GinMode: "release", FrontendURL: "https://app.example.com"})
	assert.Equal(t, []string{"https://app.example.com"}, got, "configured origin")
	got = allowedOrigins(&config.Config{})
	assert.Equal(t, []string{"http://localhost:3000"}, got, "default origin")
}
