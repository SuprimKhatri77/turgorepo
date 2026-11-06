package routes

import (
	"github.com/gin-gonic/gin"
	authHandler "github.com/suprimkhatri77/turgorepo/api/internal/handlers/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
)

func setupAuthRoutes(router *gin.RouterGroup, cfg config.Config) {
	auth := router.Group("/auth")
	auth.POST("/login",
		ratelimit.ByIP(cfg.Redis, ratelimit.LoginPerIP),
		authHandler.Login(cfg.Queries, cfg.Config, cfg.Redis),
	)
	accountEmail := authHandler.AccountEmailDeps{
		Queries: cfg.Queries,
		Config:  cfg.Config,
		Limiter: cfg.Redis,
		Revoker: cfg.Redis,
		Mailer:  cfg.Mailer,
		Jobs:    cfg.Jobs,
	}
	auth.POST("/register",
		ratelimit.ByIP(cfg.Redis, ratelimit.RegisterPerIP),
		authHandler.Register(accountEmail),
	)
	auth.POST("/verify-email",
		ratelimit.ByIP(cfg.Redis, ratelimit.VerifyEmailPerIP),
		authHandler.VerifyEmail(accountEmail),
	)
	auth.POST("/resend-verification",
		ratelimit.ByIP(cfg.Redis, ratelimit.ResendVerificationPerIP),
		authHandler.ResendVerification(accountEmail),
	)
	auth.POST("/forgot-password",
		ratelimit.ByIP(cfg.Redis, ratelimit.ForgotPasswordPerIP),
		authHandler.ForgotPassword(accountEmail),
	)
	auth.POST("/reset-password",
		ratelimit.ByIP(cfg.Redis, ratelimit.ResetPasswordPerIP),
		authHandler.ResetPassword(accountEmail),
	)

	auth.POST("/logout", authHandler.Logout(cfg.Queries, cfg.Config, cfg.Redis))
	auth.POST("/refresh", authHandler.Refresh(cfg.Queries, cfg.Config, cfg.Redis))
	auth.GET("/me", middleware.RequireAuth(cfg.Config, cfg.Redis), authHandler.Me(cfg.Queries))
}
