package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

// testTx returns a transaction on the database in TEST_DATABASE_URL, rolled
// back when the test ends so it leaves no rows behind. InTx calls on it run as
// savepoints. Without TEST_DATABASE_URL the test is skipped; CI sets it in the
// migrations job, after migrating.
func testTx(t *testing.T) pgx.Tx {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return tx
}

func createUser(ctx context.Context, q *db.Queries, email string) error {
	_, err := q.CreateUser(ctx, db.CreateUserParams{
		Name:         "Ada Lovelace",
		Email:        email,
		PasswordHash: "not-a-real-hash",
		Role:         "member",
	})
	return err
}

// userExists reports whether email is visible to q.
func userExists(t *testing.T, q *db.Queries, email string) bool {
	t.Helper()
	_, err := q.GetUserByEmail(context.Background(), email)
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	require.NoError(t, err)
	return true
}

func TestInTx(t *testing.T) {
	ctx := context.Background()
	outer := testTx(t)
	q := db.New(outer)

	t.Run("commits when fn returns nil", func(t *testing.T) {
		err := InTx(ctx, outer, func(q *db.Queries) error {
			return createUser(ctx, q, "commit@tx.test")
		})
		require.NoError(t, err)
		assert.True(t, userExists(t, q, "commit@tx.test"))
	})

	t.Run("rolls back every query when fn returns an error", func(t *testing.T) {
		failed := errors.New("second step failed")
		err := InTx(ctx, outer, func(q *db.Queries) error {
			if err := createUser(ctx, q, "rollback@tx.test"); err != nil {
				return err
			}
			return failed
		})
		require.ErrorIs(t, err, failed)
		assert.False(t, userExists(t, q, "rollback@tx.test"))
	})

	t.Run("rolls back when fn panics", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = InTx(ctx, outer, func(q *db.Queries) error {
				if err := createUser(ctx, q, "panic@tx.test"); err != nil {
					return err
				}
				panic("boom")
			})
		})
		assert.False(t, userExists(t, q, "panic@tx.test"))
	})

	t.Run("a failed query fails the whole transaction", func(t *testing.T) {
		err := InTx(ctx, outer, func(q *db.Queries) error {
			if err := createUser(ctx, q, "first@tx.test"); err != nil {
				return err
			}
			// The email is already taken, so Postgres rejects this one.
			return createUser(ctx, q, "commit@tx.test")
		})
		assert.True(t, IsUniqueViolation(err))
		assert.False(t, userExists(t, q, "first@tx.test"))
	})
}
