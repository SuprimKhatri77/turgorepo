package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
)

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
