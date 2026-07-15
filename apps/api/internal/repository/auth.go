package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateRefreshToken(ctx context.Context, params db.CreateRefreshTokenParams) (db.RefreshToken, error)
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error)
	RevokeTokenByUserIDAndToken(ctx context.Context, params db.RevokeTokenByUserIDAndTokenParams) (pgconn.CommandTag, error)
	GetRefreshTokenByUserIDAndToken(ctx context.Context, params db.GetRefreshTokenByUserIDAndTokenParams) (db.RefreshToken, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error)
}
