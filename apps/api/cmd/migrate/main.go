// Command migrate applies the embedded database migrations.
//
//	migrate up        apply all pending migrations
//	migrate down N    roll back the last N migrations
//	migrate version   print the current version
//
// It connects with MIGRATE_DATABASE_URL, falling back to DATABASE_URL. Point it at
// a direct connection: migration locks don't work through a pooler such as Neon's.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if cfg.MigrateDatabaseURL == "" {
		return errors.New("MIGRATE_DATABASE_URL or DATABASE_URL is required")
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, driverURL(cfg.MigrateDatabaseURL))
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() {
		if sourceErr, databaseErr := m.Close(); sourceErr != nil || databaseErr != nil {
			slog.Warn("failed to close migrate", "source_err", sourceErr, "database_err", databaseErr)
		}
	}()

	command := "up"
	if len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "up":
		return report(m.Up())
	case "down":
		// Requires a count, so a forgotten argument can't roll back everything.
		if len(args) != 2 {
			return errors.New("usage: migrate down N")
		}
		steps, err := strconv.Atoi(args[1])
		if err != nil || steps < 1 {
			return fmt.Errorf("invalid step count %q", args[1])
		}
		return report(m.Steps(-steps))
	case "version":
		version, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("no migrations applied")
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("version %d (dirty: %t)\n", version, dirty)
		return nil
	default:
		return fmt.Errorf("unknown command %q (use up, down N, or version)", command)
	}
}

func report(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no change")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Println("done")
	return nil
}

// driverURL swaps the postgres scheme for the one the pgx migrate driver registers.
func driverURL(databaseURL string) string {
	_, rest, found := strings.Cut(databaseURL, "://")
	if !found {
		return databaseURL
	}
	return "pgx5://" + rest
}
