package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/suprimkhatri77/turgorepo/api/internal/app"
)

func main() {
	ctx := context.Background()

	if err := app.Seed(ctx); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}
}
