package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	config := cors.DefaultConfig()

	allowOrigins := []string{"http://localhost:3000"}
	if cfg.FrontendURL != "" {
		allowOrigins = append(allowOrigins, cfg.FrontendURL)
	}

	config.AllowOrigins = allowOrigins

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
