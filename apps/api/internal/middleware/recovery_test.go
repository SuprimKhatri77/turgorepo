package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/", Recovery(), func(*gin.Context) { panic(fmt.Errorf("boom")) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "boom", "response leaks the panic")
}
