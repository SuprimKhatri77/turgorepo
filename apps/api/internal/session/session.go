package session

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

// SessionCreator stores a new session. *db.Queries implements it.
type SessionCreator interface {
	CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error)
}

func Issue(c *gin.Context, queries SessionCreator, cfg *config.Config, user db.User) error {
	sessionID := uuid.New()
	familyID := pgtype.UUID{Bytes: sessionID, Valid: true}

	tokens, err := NewTokens(cfg, user, familyID)
	if err != nil {
		return err
	}

	_, err = queries.CreateSession(c.Request.Context(), db.CreateSessionParams{
		ID:               familyID,
		UserID:           user.ID,
		CurrentTokenHash: HashToken(tokens.RefreshToken),
		UserAgent:        pgtype.Text{String: c.Request.UserAgent(), Valid: true},
		IpAddress:        pgtype.Text{String: c.ClientIP(), Valid: true},
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(cfg.RefreshTokenTTL),
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	SetCookies(c, cfg, tokens.AccessToken, tokens.RefreshToken)
	return nil
}
