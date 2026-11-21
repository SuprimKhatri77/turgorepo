package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

// Cookie names. The web app reads the same names.
const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	// LoggedInCookie is the only one scripts can read. It holds no secret and
	// lets the web app tell whether a session exists.
	LoggedInCookie = "is_logged_in"
)

// SetCookies sets the cookies for a new or rotated session.
func SetCookies(c *gin.Context, cfg *config.Config, accessToken, refreshToken string) {
	setCookie(c, cfg, AccessTokenCookie, accessToken, cfg.AccessCookieMaxAge(), true)
	setCookie(c, cfg, RefreshTokenCookie, refreshToken, cfg.RefreshCookieMaxAge(), true)
	setCookie(c, cfg, LoggedInCookie, "true", cfg.RefreshCookieMaxAge(), false)
}

// SetAccessCookie replaces only the access token, leaving the refresh token as is.
func SetAccessCookie(c *gin.Context, cfg *config.Config, accessToken string) {
	setCookie(c, cfg, AccessTokenCookie, accessToken, cfg.AccessCookieMaxAge(), true)
}

// ClearCookies removes the session cookies.
func ClearCookies(c *gin.Context, cfg *config.Config) {
	setCookie(c, cfg, AccessTokenCookie, "", -1, true)
	setCookie(c, cfg, RefreshTokenCookie, "", -1, true)
	setCookie(c, cfg, LoggedInCookie, "", -1, false)
}

// setCookie sets a cookie for the whole site. In release mode it's Secure and
// scoped to CookieDomain; locally it's host-only so it works on localhost.
func setCookie(c *gin.Context, cfg *config.Config, name, value string, maxAge int, httpOnly bool) {
	secure := cfg.IsRelease()
	domain := ""
	if secure {
		domain = cfg.CookieDomain
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", domain, secure, httpOnly)
}
