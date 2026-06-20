package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/suprimkhatri77/turgorepo/api/internal/app"
)

func main() {
	ctx := context.Background()

	a, err := app.New(ctx)
	if err != nil {
		slog.Error("startup failed", "err", err)
		os.Exit(1)
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		slog.Error("server run failed", "err", err)
		os.Exit(1)
	}
}
