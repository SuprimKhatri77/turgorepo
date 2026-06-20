package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration from environment.
type Config struct {
	Port                int
	GinMode             string
	DatabaseURL         string
	JWTRefreshSecret    string
	JWTAccessSecret     string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	BootstrapSecret     string
	FrontendUrl         string
	CookieDomain        string
	OpenAPIPath         string
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
	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	cloudinaryCloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	cloudinaryAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudinaryAPISecret := os.Getenv("CLOUDINARY_API_SECRET")
	bootstrapSecret := os.Getenv("BOOTSTRAP_SECRET")
	frontendUrl := os.Getenv("FRONTEND_URL")
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	openapiPath := os.Getenv("OPENAPI_PATH")

	return &Config{
		Port:                port,
		GinMode:             ginMode,
		DatabaseURL:         dbURL,
		JWTRefreshSecret:    jwtRefreshSecret,
		JWTAccessSecret:     jwtAccessSecret,
		CloudinaryCloudName: cloudinaryCloudName,
		CloudinaryAPIKey:    cloudinaryAPIKey,
		CloudinaryAPISecret: cloudinaryAPISecret,
		BootstrapSecret:     bootstrapSecret,
		FrontendUrl:         frontendUrl,
		CookieDomain:        cookieDomain,
		OpenAPIPath:         openapiPath,
	}, nil
}
