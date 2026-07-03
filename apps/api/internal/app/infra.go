package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/database"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
)

func initDB(ctx context.Context, cfg *config.Config) (*database.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}
	db, err := database.ConnectWithRetry(ctx, cfg.DatabaseURL, 10)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	return db, nil
}

func initCloudinary(cfg *config.Config) (*cloudinary.Client, error) {
	c, err := cloudinary.New(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: %w", err)
	}
	return c, nil
}

func initLogger(cfg *config.Config) {
	var handler slog.Handler
	if os.Getenv("GO_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	}
	slog.SetDefault(slog.New(handler))
}
