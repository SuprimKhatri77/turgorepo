package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	robfigcron "github.com/robfig/cron/v3"

	"github.com/suprimkhatri77/turgorepo/api/internal/background"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/cron"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
	apiredis "github.com/suprimkhatri77/turgorepo/api/internal/packages/redis"
	"github.com/suprimkhatri77/turgorepo/api/internal/validator"
)

const jobsShutdownTimeout = 10 * time.Second

type App struct {
	Cfg       *config.Config
	Queries   *dbgen.Queries
	DB        *database.DB
	Redis     *apiredis.Client
	CldClient *cloudinary.Client
	Router    *gin.Engine
	Scheduler *robfigcron.Cron
	Jobs      *background.Runner
}

func New(ctx context.Context) (_ *App, err error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	initLogger(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	a := &App{Cfg: cfg}
	// On failure, release whatever was set up before it. Close skips the rest.
	defer func() {
		if err != nil {
			a.Close()
		}
	}()

	if a.DB, err = initDB(ctx, cfg); err != nil {
		return nil, err
	}
	if a.Redis, err = initRedis(ctx, cfg); err != nil {
		return nil, err
	}
	if a.CldClient, err = initCloudinary(cfg); err != nil {
		return nil, err
	}

	a.Queries = dbgen.New(a.DB.Pool)
	validator.Init()

	if a.Scheduler, err = cron.Start(a.Queries); err != nil {
		return nil, fmt.Errorf("cron: %w", err)
	}

	a.Jobs = &background.Runner{}

	if a.Router, err = buildRouter(cfg, a.Queries, a.CldClient, a.DB, a.Redis, initMailer(cfg), a.Jobs); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Close() {
	if a.Scheduler != nil {
		<-a.Scheduler.Stop().Done()
	}
	// Jobs use Redis and the database, so they finish before those close.
	if a.Jobs != nil {
		ctx, cancel := context.WithTimeout(context.Background(), jobsShutdownTimeout)
		defer cancel()
		if !a.Jobs.Wait(ctx) {
			slog.Warn("shut down before background jobs finished")
		}
	}
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			slog.Warn("failed to close redis", "error", err)
		}
	}
	if a.DB != nil {
		a.DB.Close()
	}
}
