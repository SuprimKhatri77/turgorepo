package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/background"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	apiredis "github.com/suprimkhatri77/turgorepo/api/internal/packages/redis"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes"
	routesconfig "github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
)

// maxRequestBodyBytes applies to every route. If you add file uploads, move
// BodyLimit onto a route group so the upload route can have its own limit.
const maxRequestBodyBytes = 1 << 20

func buildRouter(
	cfg *config.Config,
	queries *dbgen.Queries,
	cld *cloudinary.Client,
	db *database.DB,
	redisClient *apiredis.Client,
	mailer email.Mailer,
	jobs *background.Runner,
) (*gin.Engine, error) {
	r := gin.New()

	// Without this, gin trusts every proxy and clients can fake their IP with
	// X-Forwarded-For, which would defeat per-IP rate limits.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("config: invalid TRUSTED_PROXIES: %w", err)
	}

	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	if cfg.IsRelease() {
		r.Use(middleware.RequestLogger())
	} else {
		r.Use(gin.Logger())
	}
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.BodyLimit(maxRequestBodyBytes))

	routes.Setup(r, buildRouteConfig(cfg, queries, cld, db, redisClient, mailer, jobs))
	return r, nil
}

func buildRouteConfig(
	cfg *config.Config,
	queries *dbgen.Queries,
	cld *cloudinary.Client,
	db *database.DB,
	redisClient *apiredis.Client,
	mailer email.Mailer,
	jobs *background.Runner,
) routesconfig.Config {
	return routesconfig.Config{
		Config:    cfg,
		Queries:   queries,
		CldClient: cld,
		PgxPool:   db.Pool,
		Redis:     redisClient,
		Mailer:    mailer,
		Jobs:      jobs,
	}
}
