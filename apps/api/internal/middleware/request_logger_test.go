package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLoggerSkipsPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	r := gin.New()
	r.Use(RequestLogger("/health"))
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/users", func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, path := range []string{"/health", "/users"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	assert.Contains(t, logs.String(), "path=/users")
	assert.NotContains(t, logs.String(), "path=/health")
	assert.Equal(t, 1, bytes.Count(logs.Bytes(), []byte("request completed")))
}
