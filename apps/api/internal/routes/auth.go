package routes

import (
	"github.com/gin-gonic/gin"
	authHandler "github.com/suprimkhatri77/turgorepo/api/internal/handlers/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
)

func setupAuthRoutes(router *gin.RouterGroup, deps Deps) {
	auth := router.Group("/auth")
	auth.POST("/login",
		ratelimit.ByIP(deps.Redis, ratelimit.LoginPerIP),
		authHandler.Login(deps.Queries, deps.Config, deps.Redis),
	)
	accountEmail := authHandler.AccountEmailDeps{
		Queries: deps.Queries,
		Config:  deps.Config,
		Limiter: deps.Redis,
		Revoker: deps.Redis,
		Mailer:  deps.Mailer,
		Jobs:    deps.Jobs,
	}
	auth.POST("/register",
		ratelimit.ByIP(deps.Redis, ratelimit.RegisterPerIP),
		authHandler.Register(accountEmail),
	)
	auth.POST("/verify-email",
		ratelimit.ByIP(deps.Redis, ratelimit.VerifyEmailPerIP),
		authHandler.VerifyEmail(accountEmail),
	)
	auth.POST("/resend-verification",
		ratelimit.ByIP(deps.Redis, ratelimit.ResendVerificationPerIP),
		authHandler.ResendVerification(accountEmail),
	)
	auth.POST("/forgot-password",
		ratelimit.ByIP(deps.Redis, ratelimit.ForgotPasswordPerIP),
		authHandler.ForgotPassword(accountEmail),
	)
	auth.POST("/reset-password",
		ratelimit.ByIP(deps.Redis, ratelimit.ResetPasswordPerIP),
		authHandler.ResetPassword(accountEmail),
	)

	auth.POST("/logout", authHandler.Logout(deps.Queries, deps.Config, deps.Redis))
	auth.POST("/refresh", authHandler.Refresh(deps.Queries, deps.Config, deps.Redis))
	auth.POST("/logout-all",
		middleware.RequireAuth(deps.Config, deps.Redis),
		authHandler.LogoutAll(deps.Queries, deps.Config, deps.Redis),
	)
	auth.GET("/me", middleware.RequireAuth(deps.Config, deps.Redis), authHandler.Me(deps.Queries))
}
