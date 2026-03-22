package routes

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/suprimkhatri77/turgorepo/backend/internal/handlers"
	"github.com/suprimkhatri77/turgorepo/backend/internal/repository"
)

// Config holds dependencies for route setup.
type Config struct {
	TodoRepo    *repository.TodoRepository
	OpenAPIPath string // path to openapi.json file
}

// Setup attaches all routes to the given engine.
func Setup(r *gin.Engine, cfg Config) {
	todoHandler := handlers.NewTodoHandler(cfg.TodoRepo)

	// API v1
	api := r.Group("/api")
	{
		todos := api.Group("/todos")
		{
			todos.GET("", todoHandler.List)
			todos.GET("/:id", todoHandler.GetByID)
			todos.POST("", todoHandler.Create)
			todos.PUT("/:id", todoHandler.Update)
			todos.DELETE("/:id", todoHandler.Delete)
		}
	}

	// OpenAPI spec (from generated file)
	r.GET("/openapi.json", func(c *gin.Context) {
		if cfg.OpenAPIPath == "" {
			c.Status(http.StatusNotFound)
			return
		}
		data, err := os.ReadFile(cfg.OpenAPIPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "openapi spec not found"})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	})

	// Scalar API docs UI
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/")
	})
	r.GET("/docs/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, scalarHTML)
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to the API"})
	})
}

// scalarHTML is the Scalar API docs page that loads /openapi.json.
const scalarHTML = `<!DOCTYPE html>
<html>
<head>
  <title>API Docs</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <link rel="icon" href="https://cdn.jsdelivr.net/npm/@scalar/api-reference/favicon.ico" />
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@scalar/api-reference/style.css" />
</head>
<body>
  <script id="api-reference" data-url="/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>
`
