package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	Name     string
	Email    string
	Password string
	Role     string
}

// Seed inserts demo users (idempotent). Used by cmd/seed.
func Seed(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	initLogger(cfg)

	db, err := initDB(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	queries := dbgen.New(db.Pool)

	users := []seedUser{
		{
			Name:     envOr("SEED_ADMIN_NAME", "Admin User"),
			Email:    envOr("SEED_ADMIN_EMAIL", "admin@example.com"),
			Password: envOr("SEED_ADMIN_PASSWORD", "changeme"),
			Role:     "admin",
		},
		{
			Name:     envOr("SEED_MEMBER_NAME", "Member User"),
			Email:    envOr("SEED_MEMBER_EMAIL", "member@example.com"),
			Password: envOr("SEED_MEMBER_PASSWORD", "changeme"),
			Role:     "member",
		},
	}

	for _, u := range users {
		if err := upsertSeedUser(ctx, queries, u); err != nil {
			return err
		}
	}

	slog.Info("seed complete",
		"admin_email", users[0].Email,
		"member_email", users[1].Email,
	)
	return nil
}

func upsertSeedUser(ctx context.Context, queries *dbgen.Queries, u seedUser) error {
	existing, err := queries.GetUserByEmail(ctx, u.Email)
	if err == nil {
		slog.Info("seed user already exists, skipping", "email", existing.Email, "role", existing.Role)
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("lookup %s: %w", u.Email, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password for %s: %w", u.Email, err)
	}

	created, err := queries.CreateUser(ctx, dbgen.CreateUserParams{
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: string(hash),
		Role:         u.Role,
		ImageUrl:     pgtype.Text{Valid: false},
	})
	if err != nil {
		return fmt.Errorf("create %s: %w", u.Email, err)
	}

	slog.Info("seeded user", "email", created.Email, "role", created.Role, "id", created.ID)
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
