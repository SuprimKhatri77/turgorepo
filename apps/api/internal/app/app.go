package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
	apiredis "github.com/suprimkhatri77/turgorepo/api/internal/packages/redis"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

type App struct {
	Cfg       *config.Config
	Queries   *dbgen.Queries
	DB        *database.DB
	Redis     *apiredis.Client
	CldClient *cloudinary.Client
	Router    *gin.Engine
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	initLogger(cfg)

	db, err := initDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	redisClient, err := initRedis(ctx, cfg)
	if err != nil {
		db.Close()
		return nil, err
	}

	cldClient, err := initCloudinary(cfg)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, err
	}

	queries := dbgen.New(db.Pool)
	validator.Init()

	// Initialize cron jobs when needed:
	// cron.CronExample(queries)

	r := buildRouter(cfg, queries, cldClient, db, redisClient)

	return &App{
		Cfg:       cfg,
		Queries:   queries,
		DB:        db,
		Redis:     redisClient,
		CldClient: cldClient,
		Router:    r,
	}, nil
}

func (a *App) Close() {
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			slog.Warn("failed to close redis", "err", err)
		}
	}
	a.DB.Close()
}
