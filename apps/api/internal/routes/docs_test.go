package routes

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
)

func TestDocsPageNonce(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	setupDocsRoutes(r.Group("/api/v1"), Deps{Config: &config.Config{}})

	nonceAttr := regexp.MustCompile(`nonce="([^"]+)"`)
	get := func() (string, string) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/", nil))
		require.Equal(t, http.StatusOK, w.Code)
		matches := nonceAttr.FindAllStringSubmatch(w.Body.String(), -1)
		require.Len(t, matches, 2)
		require.Equal(t, matches[0][1], matches[1][1], "script nonces")
		return matches[0][1], w.Header().Get("Content-Security-Policy")
	}

	nonce, policy := get()
	assert.Contains(t, policy, "script-src 'nonce-"+nonce+"' 'strict-dynamic'", "policy doesn't allow the nonce")
	assert.Contains(t, policy, "frame-ancestors 'none'", "policy allows framing")
	again, _ := get()
	assert.NotEqual(t, nonce, again, "nonce reused across requests")
}
