package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	config := cors.DefaultConfig()

	config.AllowOrigins = allowedOrigins(cfg)

	config.AllowCredentials = true

	config.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD",
	}

	config.AllowHeaders = []string{
		"Origin", "Content-Length", "Content-Type",
		"Authorization", "Accept", "X-Request-ID",
	}

	config.ExposeHeaders = []string{"Content-Length", "X-Request-ID"}

	return cors.New(config)
}

// allowedOrigins never includes localhost in release mode, where credentialed
// requests from a developer's machine must not reach production.
func allowedOrigins(cfg *config.Config) []string {
	if cfg.FrontendURL != "" {
		return []string{cfg.FrontendURL}
	}
	return []string{"http://localhost:3000"}
}
