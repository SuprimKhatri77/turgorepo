package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration from environment.
type Config struct {
	Port                int
	GinMode             string
	DatabaseURL         string
	MigrateDatabaseURL  string
	RedisURL            string
	JWTRefreshSecret    string
	JWTAccessSecret     string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	FrontendURL         string
	CookieDomain        string
	OpenAPIPath         string
	TrustedProxies      []string
	DocsEnabled         bool
	SMTPHost            string
	SMTPPort            int
	SMTPUsername        string
	SMTPPassword        string
	SMTPAllowPlaintext  bool
	EmailFrom           string
	PasswordResetTTL    time.Duration
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	RefreshReuseWindow  time.Duration
}

// Load loads .env from the current directory (if present) then reads configuration from environment variables.
// No hardcoded URLs: PORT defaults to 5000 if unset; GIN_MODE defaults to "debug"; DATABASE_URL must be set in .env or env.
func Load() (*Config, error) {
	_ = godotenv.Load("../../.env.local")

	port := 5000
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	cloudinaryCloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	cloudinaryAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudinaryAPISecret := os.Getenv("CLOUDINARY_API_SECRET")
	frontendURL := os.Getenv("FRONTEND_URL")
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	openapiPath := os.Getenv("OPENAPI_PATH")

	return &Config{
		Port:                port,
		GinMode:             ginMode,
		DatabaseURL:         dbURL,
		MigrateDatabaseURL:  firstNonEmpty(os.Getenv("MIGRATE_DATABASE_URL"), dbURL),
		RedisURL:            redisURL,
		JWTRefreshSecret:    jwtRefreshSecret,
		JWTAccessSecret:     jwtAccessSecret,
		CloudinaryCloudName: cloudinaryCloudName,
		CloudinaryAPIKey:    cloudinaryAPIKey,
		CloudinaryAPISecret: cloudinaryAPISecret,
		FrontendURL:         frontendURL,
		CookieDomain:        cookieDomain,
		OpenAPIPath:         openapiPath,
		TrustedProxies:      splitList(os.Getenv("TRUSTED_PROXIES")),
		DocsEnabled:         docsEnabled(os.Getenv("DOCS_ENABLED"), ginMode),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            smtpPort(os.Getenv("SMTP_PORT")),
		SMTPUsername:        os.Getenv("SMTP_USERNAME"),
		SMTPPassword:        os.Getenv("SMTP_PASSWORD"),
		SMTPAllowPlaintext:  smtpAllowPlaintext(os.Getenv("SMTP_ALLOW_PLAINTEXT"), ginMode),
		EmailFrom:           os.Getenv("EMAIL_FROM"),
		PasswordResetTTL:    time.Hour,
		AccessTokenTTL:      15 * time.Minute,
		RefreshTokenTTL:     30 * 24 * time.Hour,
		RefreshReuseWindow:  5 * time.Minute,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// docsEnabled keeps API docs off in release mode unless explicitly turned on.
func docsEnabled(value, ginMode string) bool {
	if enabled, err := strconv.ParseBool(value); err == nil {
		return enabled
	}
	return ginMode != "release"
}

func smtpPort(value string) int {
	if port, err := strconv.Atoi(value); err == nil {
		return port
	}
	return defaultSMTPPort
}

// smtpAllowPlaintext requires TLS in release mode unless explicitly turned off,
// which is only meant for testing production images against a local mail server.
func smtpAllowPlaintext(value, ginMode string) bool {
	if allowed, err := strconv.ParseBool(value); err == nil {
		return allowed
	}
	return ginMode != "release"
}

// splitList parses a comma-separated env value, dropping blanks.
func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

const (
	minJWTSecretBytes = 32
	defaultSMTPPort   = 587
)

// IsRelease reports whether the API is running in production mode.
func (c *Config) IsRelease() bool {
	return c.GinMode == "release"
}

// Validate rejects configuration the API must not start with. JWT secrets are
// always required; release mode also rejects weak secrets, since anyone who can
// guess one can forge tokens for any user.
func (c *Config) Validate() error {
	var errs []error

	if c.JWTAccessSecret == "" {
		errs = append(errs, errors.New("JWT_ACCESS_SECRET is required"))
	}
	if c.JWTRefreshSecret == "" {
		errs = append(errs, errors.New("JWT_REFRESH_SECRET is required"))
	}

	if c.SMTPHost != "" && c.EmailFrom == "" {
		errs = append(errs, errors.New("EMAIL_FROM is required when SMTP_HOST is set"))
	}

	if c.IsRelease() {
		errs = append(errs, validateReleaseSecret("JWT_ACCESS_SECRET", c.JWTAccessSecret))
		errs = append(errs, validateReleaseSecret("JWT_REFRESH_SECRET", c.JWTRefreshSecret))
		if c.JWTAccessSecret != "" && c.JWTAccessSecret == c.JWTRefreshSecret {
			errs = append(errs, errors.New("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be different"))
		}
		if c.FrontendURL == "" {
			errs = append(errs, errors.New("FRONTEND_URL is required in release mode"))
		}
		if c.CookieDomain == "" {
			errs = append(errs, errors.New("COOKIE_DOMAIN is required in release mode"))
		}
		// Without email, people silently can't reset their passwords.
		if c.SMTPHost == "" {
			errs = append(errs, errors.New("SMTP_HOST is required in release mode"))
		}
	}

	return errors.Join(errs...)
}

func validateReleaseSecret(name, secret string) error {
	if secret == "" {
		return nil // already reported as missing
	}
	if len(secret) < minJWTSecretBytes {
		return fmt.Errorf("%s must be at least %d bytes in release mode", name, minJWTSecretBytes)
	}
	if strings.Contains(strings.ToLower(secret), "change-me") {
		return fmt.Errorf("%s is still a placeholder value", name)
	}
	return nil
}

func (c *Config) AccessCookieMaxAge() int {
	return int(c.AccessTokenTTL.Seconds())
}

func (c *Config) RefreshCookieMaxAge() int {
	return int(c.RefreshTokenTTL.Seconds())
}
