package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeLimiter struct {
	allowed    bool
	retryAfter time.Duration
	err        error
	key        string
}

func (l *fakeLimiter) Allow(_ context.Context, key string, _ int, _ time.Duration) (bool, time.Duration, error) {
	l.key = key
	return l.allowed, l.retryAfter, l.err
}

func serve(limiter Limiter) (*httptest.ResponseRecorder, *bool) {
	gin.SetMode(gin.TestMode)
	called := false
	r := gin.New()
	r.POST("/", ByIP(limiter, LoginPerIP), func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "203.0.113.7:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, &called
}

func TestByIP(t *testing.T) {
	t.Run("allows within budget", func(t *testing.T) {
		limiter := &fakeLimiter{allowed: true}
		w, called := serve(limiter)
		if w.Code != http.StatusOK || !*called {
			t.Errorf("status = %d, called = %v", w.Code, *called)
		}
		if limiter.key != "ratelimit:login:ip:203.0.113.7" {
			t.Errorf("key = %q", limiter.key)
		}
	})

	t.Run("blocks over budget and stops the chain", func(t *testing.T) {
		w, called := serve(&fakeLimiter{retryAfter: 1500 * time.Millisecond})
		if w.Code != http.StatusTooManyRequests || *called {
			t.Fatalf("status = %d, called = %v", w.Code, *called)
		}
		if got := w.Header().Get("Retry-After"); got != "2" {
			t.Errorf("Retry-After = %q", got)
		}
	})

	t.Run("limiter failure lets the request through", func(t *testing.T) {
		w, called := serve(&fakeLimiter{err: errors.New("redis down")})
		if w.Code != http.StatusOK || !*called {
			t.Errorf("status = %d, called = %v", w.Code, *called)
		}
	})
}
