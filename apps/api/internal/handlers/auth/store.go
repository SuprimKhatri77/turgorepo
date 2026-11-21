package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

// Store is the database access the auth handlers need. *db.Queries implements
// it, and tests use fakes.
type Store interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error)

	CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error)
	GetActiveSessionByID(ctx context.Context, id pgtype.UUID) (db.Session, error)
	RotateSessionToken(ctx context.Context, params db.RotateSessionTokenParams) (db.Session, error)
	RevokeSession(ctx context.Context, id pgtype.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID pgtype.UUID) error

	CreatePasswordResetToken(ctx context.Context, params db.CreatePasswordResetTokenParams) error
	ResetPasswordWithToken(ctx context.Context, params db.ResetPasswordWithTokenParams) (db.User, error)

	CreateEmailVerificationToken(ctx context.Context, params db.CreateEmailVerificationTokenParams) error
	VerifyEmailWithToken(ctx context.Context, tokenHash string) (db.User, error)
}

var _ Store = (*db.Queries)(nil)
