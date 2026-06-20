package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/middleware"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
	"github.com/suprimkhatri77/turgorepo/api/internal/routes"
	routesconfig "github.com/suprimkhatri77/turgorepo/api/internal/routes/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

type App struct {
	Cfg       *config.Config
	Queries   *dbgen.Queries
	DB        *database.DB
	CldClient *cloudinary.Client
	Router    *gin.Engine
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required (set it in .env or environment)")
	}

	gin.SetMode(cfg.GinMode)

	cldClient, err := cloudinary.New(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: %w", err)
	}

	db, err := database.ConnectWithRetry(ctx, cfg.DatabaseURL, 10)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	queries := dbgen.New(db.Pool)
	validator.Init()

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(cfg))

	routes.Setup(r, routesconfig.Config{
		Config:    cfg,
		Queries:   queries,
		CldClient: cldClient,
		PgxPool:   db.Pool,
	})

	return &App{Cfg: cfg, Queries: queries, DB: db, CldClient: cldClient, Router: r}, nil
}

func (a *App) Close() {
	a.DB.Close()
}
