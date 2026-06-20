package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

func Setup(r *gin.Engine, cfg config.Config) {
	router := r.Group("/api/v1")

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.APIResponse{
			Success: true,
			Message: "Server is up and running",
		})
	})

	setupAuthRoutes(router, cfg)
	setupDocsRoutes(router, cfg)
}
