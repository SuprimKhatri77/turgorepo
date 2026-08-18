package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
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
// Passwords must be set via SEED_ADMIN_PASSWORD and SEED_MEMBER_PASSWORD.
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

	adminPassword, err := requireEnv("SEED_ADMIN_PASSWORD")
	if err != nil {
		return err
	}
	memberPassword, err := requireEnv("SEED_MEMBER_PASSWORD")
	if err != nil {
		return err
	}

	users := []seedUser{
		{
			Name:     envOr("SEED_ADMIN_NAME", "Admin User"),
			Email:    envOr("SEED_ADMIN_EMAIL", "admin@example.com"),
			Password: adminPassword,
			Role:     constants.RoleAdmin,
		},
		{
			Name:     envOr("SEED_MEMBER_NAME", "Member User"),
			Email:    envOr("SEED_MEMBER_EMAIL", "member@example.com"),
			Password: memberPassword,
			Role:     constants.RoleMember,
		},
	}

	if strings.EqualFold(users[0].Email, users[1].Email) {
		return fmt.Errorf("SEED_ADMIN_EMAIL and SEED_MEMBER_EMAIL must be different")
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Info("seed user already exists (concurrent), skipping", "email", u.Email)
			return nil
		}
		return fmt.Errorf("create %s: %w", u.Email, err)
	}

	slog.Info("seeded user", "email", created.Email, "role", created.Role, "id", created.ID)
	return nil
}

func requireEnv(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("%s is required (set it in .env.local before running seed)", key)
	}
	return v, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
