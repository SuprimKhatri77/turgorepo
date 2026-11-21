package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	strongAccess  = "0123456789abcdef0123456789abcdef-access"
	strongRefresh = "0123456789abcdef0123456789abcdef-refresh"
)

func releaseConfig() Config {
	return Config{
		GinMode:          "release",
		JWTAccessSecret:  strongAccess,
		JWTRefreshSecret: strongRefresh,
		FrontendURL:      "https://app.example.com",
		CookieDomain:     ".example.com",
		SMTPHost:         "email-smtp.us-east-1.amazonaws.com",
		EmailFrom:        "no-reply@example.com",
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string
	}{
		{name: "valid release config", modify: func(*Config) {}},
		{name: "empty access secret", modify: func(c *Config) { c.JWTAccessSecret = "" }, wantErr: "JWT_ACCESS_SECRET is required"},
		{name: "empty refresh secret", modify: func(c *Config) { c.JWTRefreshSecret = "" }, wantErr: "JWT_REFRESH_SECRET is required"},
		{name: "short secret", modify: func(c *Config) { c.JWTAccessSecret = "access-secret" }, wantErr: "at least 32 bytes"},
		{name: "placeholder secret", modify: func(c *Config) { c.JWTRefreshSecret = "change-me-change-me-change-me-change-me" }, wantErr: "placeholder"},
		{name: "same secrets", modify: func(c *Config) { c.JWTRefreshSecret = strongAccess }, wantErr: "must be different"},
		{name: "no frontend url", modify: func(c *Config) { c.FrontendURL = "" }, wantErr: "FRONTEND_URL"},
		{name: "no cookie domain", modify: func(c *Config) { c.CookieDomain = "" }, wantErr: "COOKIE_DOMAIN"},
		{name: "no smtp host", modify: func(c *Config) { c.SMTPHost = "" }, wantErr: "SMTP_HOST is required"},
		{
			name:    "plaintext with a real mail server",
			modify:  func(c *Config) { c.SMTPAllowPlaintext = true },
			wantErr: "SMTP_ALLOW_PLAINTEXT",
		},
		{
			name: "plaintext with a local mail server",
			modify: func(c *Config) {
				c.SMTPAllowPlaintext = true
				c.SMTPHost = "mailpit"
			},
		},
		{
			name: "debug mode allows plaintext anywhere",
			modify: func(c *Config) {
				c.GinMode = "debug"
				c.SMTPAllowPlaintext = true
			},
		},
		{name: "no email from", modify: func(c *Config) { c.EmailFrom = "" }, wantErr: "EMAIL_FROM"},
		{
			name: "debug mode allows weak secrets",
			modify: func(c *Config) {
				c.GinMode = "debug"
				c.JWTAccessSecret = "access-secret"
				c.JWTRefreshSecret = "access-secret"
				c.FrontendURL = ""
				c.CookieDomain = ""
				c.SMTPHost = ""
				c.EmailFrom = ""
			},
		},
		{
			name: "debug mode still requires secrets",
			modify: func(c *Config) {
				c.GinMode = "debug"
				c.JWTAccessSecret = ""
			},
			wantErr: "JWT_ACCESS_SECRET is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := releaseConfig()
			tt.modify(&cfg)
			err := cfg.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestSplitList(t *testing.T) {
	got := splitList(" 10.0.0.0/16, ,192.168.0.0/24 ")
	assert.Equal(t, []string{"10.0.0.0/16", "192.168.0.0/24"}, got)
	assert.Nil(t, splitList(""), "empty value should give nil")
}

func TestDocsEnabled(t *testing.T) {
	tests := []struct {
		value, mode string
		want        bool
	}{
		{"", "debug", true},
		{"", "release", false},
		{"true", "release", true},
		{"false", "debug", false},
		{"not-a-bool", "release", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, docsEnabled(tt.value, tt.mode), "docsEnabled(%q, %q)", tt.value, tt.mode)
	}
}

func TestSMTPAllowPlaintext(t *testing.T) {
	tests := []struct {
		value, mode string
		want        bool
	}{
		{"", "debug", true},
		{"", "release", false},
		{"true", "release", true},
		{"false", "debug", false},
		{"not-a-bool", "release", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, smtpAllowPlaintext(tt.value, tt.mode), "smtpAllowPlaintext(%q, %q)", tt.value, tt.mode)
	}
}

func TestSMTPPort(t *testing.T) {
	assert.Equal(t, 1025, smtpPort("1025"))
	assert.Equal(t, defaultSMTPPort, smtpPort(""), "empty value")
}
