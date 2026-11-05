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
		AccessTokenTTL:      15 * time.Minute,
		RefreshTokenTTL:     30 * 24 * time.Hour,
		RefreshReuseWindow:  5 * time.Minute,
	}, nil
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

const minJWTSecretBytes = 32

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
