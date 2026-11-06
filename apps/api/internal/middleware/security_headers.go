package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

// APIContentSecurityPolicy suits JSON responses: a browser should never run or
// load anything from them, or show them in a frame. A route that serves a page,
// such as the docs, sets its own policy instead.
const APIContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'"

// SecurityHeaders sets headers that stop browsers from guessing content types,
// framing responses or sending the API's URLs to other sites.
func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	release := cfg.IsRelease()
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", APIContentSecurityPolicy)
		// Production runs behind HTTPS. Browsers ignore this header over plain HTTP.
		if release {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
