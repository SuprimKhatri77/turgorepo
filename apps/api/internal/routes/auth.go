package routes

import (
	"github.com/gin-gonic/gin"
	authHandler "github.com/suprimkhatri77/turgorepo/api/internal/handlers/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
)

func setupAuthRoutes(router *gin.RouterGroup, cfg config.Config) {
	auth := router.Group("/auth")
	auth.POST("/login", authHandler.Login(cfg.Queries, cfg.Config))
	auth.POST("/register", authHandler.Register(cfg.Queries, cfg.Config))
	auth.POST("/logout", authHandler.Logout(cfg.Queries, cfg.Config))
	auth.POST("/refresh", authHandler.Refresh(cfg.Queries, cfg.Config))
	auth.GET("/me", middleware.RequireAuth(cfg.Config), middleware.RequireRole("admin", "superadmin", "staff", "member"), authHandler.Me(cfg.Queries))
}
