package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	config := cors.DefaultConfig()

	allowOrigins := []string{"http://localhost:3000"}
	if cfg.FrontendUrl != "" {
		allowOrigins = append(allowOrigins, cfg.FrontendUrl)
	}

	config.AllowOrigins = allowOrigins

	config.AllowCredentials = true

	config.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD",
	}

	config.AllowHeaders = []string{
		"Origin", "Content-Length", "Content-Type",
		"Authorization", "Accept",
	}

	config.ExposeHeaders = []string{"Content-Length"}

	return cors.New(config)
}
