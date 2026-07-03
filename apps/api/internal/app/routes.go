package app

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes"
	routesconfig "github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
)

func buildRouter(cfg *config.Config, queries *dbgen.Queries, cld *cloudinary.Client, db *database.DB) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(cfg))

	routes.Setup(r, buildRouteConfig(cfg, queries, cld, db))
	return r
}

func buildRouteConfig(cfg *config.Config, queries *dbgen.Queries, cld *cloudinary.Client, db *database.DB) routesconfig.Config {
	return routesconfig.Config{
		Config:    cfg,
		Queries:   queries,
		CldClient: cld,
		PgxPool:   db.Pool,
	}
}
