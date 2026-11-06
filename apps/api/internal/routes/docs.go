package routes

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
)

func setupDocsRoutes(r *gin.RouterGroup, cfg config.Config) {
	r.GET("/openapi.json", func(c *gin.Context) {
		if cfg.Config.OpenAPIPath == "" {
			c.Status(http.StatusNotFound)
			return
		}
		data, err := os.ReadFile(cfg.Config.OpenAPIPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "openapi spec not found"})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	})

	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, r.BasePath()+"/docs/")
	})
	r.GET("/docs/", func(c *gin.Context) {
		nonce, err := newNonce()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		html := strings.Replace(scalarHTML, `data-url="/openapi.json"`, `data-url="`+r.BasePath()+`/openapi.json"`, 1)
		html = strings.ReplaceAll(html, "{{nonce}}", nonce)

		c.Header("Content-Security-Policy", docsContentSecurityPolicy(nonce))
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	})
}

// docsContentSecurityPolicy replaces the API's default policy, which allows
// nothing, with what the Scalar page needs. The browser console reports one
// blocked eval: Scalar's bundled Zod checks whether eval is allowed and works
// without it, so the policy doesn't allow it.
func docsContentSecurityPolicy(nonce string) string {
	return strings.Join([]string{
		"default-src 'none'",
		"script-src 'nonce-" + nonce + "' 'strict-dynamic'",
		"style-src 'unsafe-inline' https://fonts.scalar.com",
		"font-src https://fonts.scalar.com data:",
		"img-src 'self' data: https:",
		"connect-src 'self'",
		"base-uri 'none'",
		"form-action 'none'",
		"frame-ancestors 'none'",
	}, "; ")
}

func newNonce() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

// scalarHTML is the Scalar API docs page that loads /openapi.json.
const scalarHTML = `<!DOCTYPE html>
<html>
<head>
  <title>API Docs</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  <script id="api-reference" data-url="/openapi.json" nonce="{{nonce}}"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.77.0" nonce="{{nonce}}"></script>
</body>
</html>
`
