package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

// TxStarter begins a transaction. *pgxpool.Pool is the usual one. A pgx.Tx
// also works: it begins a savepoint, so InTx calls can nest.
type TxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// InTx runs fn in a transaction and commits if it returns nil. It rolls back if
// fn returns an error or panics, so callers never begin, commit or roll back.
//
// Use it for operations whose queries must succeed or fail together, as a
// method on the feature's store, so handlers call one method:
//
//	err := database.InTx(ctx, s.pool, func(q *db.Queries) error {
//		user, err := q.CreateUser(ctx, params)
//		if err != nil {
//			return err
//		}
//		return q.CreateWorkspace(ctx, user.ID)
//	})
//
// Run every query through q, which is bound to the transaction; queries on the
// pool would run outside it. Keep slow work (password hashing, emails, uploads)
// out of fn, since the transaction holds a connection and its locks until it
// ends. Pass the request context, so a stuck transaction is cancelled.
func InTx(ctx context.Context, conn TxStarter, fn func(q *db.Queries) error) error {
	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		return fn(db.New(tx))
	})
}
