package config

import (
	"strings"
	"testing"
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
		{
			name: "debug mode allows weak secrets",
			modify: func(c *Config) {
				c.GinMode = "debug"
				c.JWTAccessSecret = "access-secret"
				c.JWTRefreshSecret = "access-secret"
				c.FrontendURL = ""
				c.CookieDomain = ""
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
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestSplitList(t *testing.T) {
	got := splitList(" 10.0.0.0/16, ,192.168.0.0/24 ")
	if len(got) != 2 || got[0] != "10.0.0.0/16" || got[1] != "192.168.0.0/24" {
		t.Errorf("got %v", got)
	}
	if got := splitList(""); got != nil {
		t.Errorf("empty value should give nil, got %v", got)
	}
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
		if got := docsEnabled(tt.value, tt.mode); got != tt.want {
			t.Errorf("docsEnabled(%q, %q) = %v, want %v", tt.value, tt.mode, got, tt.want)
		}
	}
}
