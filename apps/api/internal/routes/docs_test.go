package routes

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	apiconfig "github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
)

func TestDocsPageNonce(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	setupDocsRoutes(r.Group("/api/v1"), config.Config{Config: &apiconfig.Config{}})

	nonceAttr := regexp.MustCompile(`nonce="([^"]+)"`)
	get := func() (string, string) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		matches := nonceAttr.FindAllStringSubmatch(w.Body.String(), -1)
		if len(matches) != 2 || matches[0][1] != matches[1][1] {
			t.Fatalf("script nonces = %v", matches)
		}
		return matches[0][1], w.Header().Get("Content-Security-Policy")
	}

	nonce, policy := get()
	if !strings.Contains(policy, "script-src 'nonce-"+nonce+"' 'strict-dynamic'") {
		t.Errorf("policy %q doesn't allow nonce %q", policy, nonce)
	}
	if !strings.Contains(policy, "frame-ancestors 'none'") {
		t.Errorf("policy %q allows framing", policy)
	}
	if again, _ := get(); again == nonce {
		t.Error("nonce reused across requests")
	}
}
