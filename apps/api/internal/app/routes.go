package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes"
)

// maxRequestBodyBytes applies to every route. If you add file uploads, move
// BodyLimit onto a route group so the upload route can have its own limit.
const maxRequestBodyBytes = 1 << 20

// buildRouter applies the global middleware and mounts the routes.
func buildRouter(deps routes.Deps) (*gin.Engine, error) {
	cfg := deps.Config
	r := gin.New()

	// Without this, gin trusts every proxy and clients can fake their IP with
	// X-Forwarded-For, which would defeat per-IP rate limits.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("config: invalid TRUSTED_PROXIES: %w", err)
	}

	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	if cfg.IsRelease() {
		r.Use(middleware.RequestLogger(routes.HealthPath))
	} else {
		r.Use(gin.Logger())
	}
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.BodyLimit(maxRequestBodyBytes))

	routes.Setup(r, deps)
	return r, nil
}
