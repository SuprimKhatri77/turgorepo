package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

func TestBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serve := func(req *http.Request) (*httptest.ResponseRecorder, *bool, *error) {
		reached := false
		var readErr error
		r := gin.New()
		r.POST("/", BodyLimit(10), func(c *gin.Context) {
			reached = true
			_, readErr = io.ReadAll(c.Request.Body)
			c.Status(http.StatusOK)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w, &reached, &readErr
	}

	t.Run("allows a body within the limit", func(t *testing.T) {
		w, reached, readErr := serve(httptest.NewRequest(http.MethodPost, "/", strings.NewReader("short")))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, *reached, "handler reached")
		assert.NoError(t, *readErr)
	})

	t.Run("refuses a declared oversized body without reading it", func(t *testing.T) {
		w, reached, _ := serve(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 11))))
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.False(t, *reached, "handler reached")
	})

	t.Run("stops reading a body that hides its size", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 11)))
		req.ContentLength = -1
		_, reached, readErr := serve(req)
		assert.True(t, *reached, "handler reached")
		assert.Error(t, *readErr)
	})
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		incoming string
		reused   bool
	}{
		{"reuses a normal id", "abc-123_DEF", true},
		{"mints when missing", "", false},
		{"mints when too long", strings.Repeat("a", maxRequestIDLength+1), false},
		{"mints when it has log-unsafe characters", `abc" level=ERROR`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", RequestID(), func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(constants.RequestIDHeader, tt.incoming)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			got := w.Header().Get(constants.RequestIDHeader)
			require.NotEmpty(t, got, "no request id on the response")
			assert.Equal(t, tt.reused, got == tt.incoming, "reused: got %q, incoming %q", got, tt.incoming)
		})
	}
}

func TestAllowedOrigins(t *testing.T) {
	got := allowedOrigins(&config.Config{GinMode: "release", FrontendURL: "https://app.example.com"})
	assert.Equal(t, []string{"https://app.example.com"}, got, "configured origin")
	got = allowedOrigins(&config.Config{})
	assert.Equal(t, []string{"http://localhost:3000"}, got, "default origin")
}

func TestRecoveryReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/", Recovery(), func(*gin.Context) { panic(fmt.Errorf("boom")) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "boom", "response leaks the panic")
}
