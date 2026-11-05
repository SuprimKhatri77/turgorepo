package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

const readyTimeout = 2 * time.Second

func Setup(r *gin.Engine, cfg config.Config) {
	router := r.Group("/api/v1")

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to the API"})
	})

	// Liveness: the process is up. Load balancers should use this one, so a brief
	// database problem doesn't make them replace every container.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Server is up and running",
		})
	})

	router.GET("/ready", ready(cfg.PgxPool, cfg.Redis))

	setupAuthRoutes(router, cfg)
	if cfg.Config.DocsEnabled {
		setupDocsRoutes(router, cfg)
	}
}

// Pinger is a dependency that /ready checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// ready reports whether the API can reach its dependencies. Failures are logged
// rather than returned, since this endpoint is public.
func ready(dependencies ...Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readyTimeout)
		defer cancel()

		for _, dependency := range dependencies {
			if err := dependency.Ping(ctx); err != nil {
				rlog.Error(c, "readiness check failed", err)
				c.JSON(http.StatusServiceUnavailable, types.APIResponse{
					Success: false,
					Message: "Service is not ready",
					Code:    constants.ServiceUnavailable,
				})
				return
			}
		}

		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Service is ready",
		})
	}
}
