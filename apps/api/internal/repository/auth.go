package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateToken(ctx context.Context, params db.CreateTokenParams) (db.Token, error)
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error)
	RevokeTokenBySessionIDAndToken(ctx context.Context, params db.RevokeTokenBySessionIDAndTokenParams) (pgconn.CommandTag, error)
	GetRefreshTokenBySessionIDAndToken(ctx context.Context, params db.GetRefreshTokenBySessionIDAndTokenParams) (db.Token, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error)
}
