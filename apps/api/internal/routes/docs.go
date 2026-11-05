package routes

import (
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
		c.Header("Content-Type", "text/html; charset=utf-8")
		html := strings.Replace(scalarHTML, `data-url="/openapi.json"`, `data-url="`+r.BasePath()+`/openapi.json"`, 1)
		c.String(http.StatusOK, html)
	})
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
  <script id="api-reference" data-url="/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.77.0"></script>
</body>
</html>
`
