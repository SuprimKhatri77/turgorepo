package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("a request that waits too long is cancelled", func(t *testing.T) {
		var waitErr error
		r := gin.New()
		r.GET("/", RequestTimeout(10*time.Millisecond), func(c *gin.Context) {
			// Stands in for a query to a database that has stopped answering.
			<-c.Request.Context().Done()
			waitErr = c.Request.Context().Err()
		})

		done := make(chan struct{})
		go func() {
			r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("request still running after the timeout")
		}
		assert.ErrorIs(t, waitErr, context.DeadlineExceeded)
	})

	t.Run("a fast request is unaffected", func(t *testing.T) {
		r := gin.New()
		r.GET("/", RequestTimeout(time.Minute), func(c *gin.Context) {
			_, hasDeadline := c.Request.Context().Deadline()
			assert.True(t, hasDeadline)
			c.Status(http.StatusOK)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
