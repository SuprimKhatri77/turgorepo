package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/handlers/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
)

func setupAuthRoutes(router *gin.RouterGroup, deps Deps) {
	h := &auth.Handler{
		Store:   deps.Queries,
		Config:  deps.Config,
		Limiter: deps.Redis,
		Revoker: deps.Redis,
		Mailer:  deps.Mailer,
		Jobs:    deps.Jobs,
	}
	requireAuth := middleware.RequireAuth(deps.Config, deps.Redis)

	r := router.Group("/auth")
	r.POST("/login", ratelimit.ByIP(deps.Redis, ratelimit.LoginPerIP), h.Login)
	r.POST("/register", ratelimit.ByIP(deps.Redis, ratelimit.RegisterPerIP), h.Register)
	r.POST("/verify-email", ratelimit.ByIP(deps.Redis, ratelimit.VerifyEmailPerIP), h.VerifyEmail)
	r.POST("/resend-verification", ratelimit.ByIP(deps.Redis, ratelimit.ResendVerificationPerIP), h.ResendVerification)
	r.POST("/forgot-password", ratelimit.ByIP(deps.Redis, ratelimit.ForgotPasswordPerIP), h.ForgotPassword)
	r.POST("/reset-password", ratelimit.ByIP(deps.Redis, ratelimit.ResetPasswordPerIP), h.ResetPassword)

	r.POST("/logout", h.Logout)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout-all", requireAuth, h.LogoutAll)
	r.GET("/me", requireAuth, h.Me)
}
