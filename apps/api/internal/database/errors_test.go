package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	assert.True(t, IsUniqueViolation(&pgconn.PgError{Code: "23505"}))
	assert.True(t, IsUniqueViolation(fmt.Errorf("create user: %w", &pgconn.PgError{Code: "23505"})), "wrapped")
	assert.False(t, IsUniqueViolation(&pgconn.PgError{Code: "23503"}), "foreign key violation")
	assert.False(t, IsUniqueViolation(errors.New("connection refused")))
	assert.False(t, IsUniqueViolation(nil))
}
