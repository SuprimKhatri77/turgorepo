package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolation is Postgres's SQLSTATE for a duplicate value in a unique
// column or index.
const uniqueViolation = "23505"

// IsUniqueViolation reports whether err is Postgres refusing a duplicate value,
// such as registering an email that's already taken.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
