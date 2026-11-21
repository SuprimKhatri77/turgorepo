package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
