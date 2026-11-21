package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serve := func(mode string) http.Header {
		r := gin.New()
		r.Use(SecurityHeaders(&config.Config{GinMode: mode}))
		r.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		return w.Header()
	}

	headers := serve("debug")
	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "no-referrer",
		"Content-Security-Policy": APIContentSecurityPolicy,
	}
	for name, value := range want {
		assert.Equal(t, value, headers.Get(name), name)
	}
	assert.Empty(t, headers.Get("Strict-Transport-Security"), "HSTS outside release mode")

	assert.Equal(t, "max-age=31536000; includeSubDomains", serve("release").Get("Strict-Transport-Security"), "HSTS in release mode")
}

func TestSecurityHeadersLetRoutesOverrideThePolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeaders(&config.Config{}))
	r.GET("/page", func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/page", nil))

	assert.Equal(t, []string{"default-src 'self'"}, w.Header().Values("Content-Security-Policy"))
}
