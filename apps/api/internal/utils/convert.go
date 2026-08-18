package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ConvertToUUID(id string) (pgtype.UUID, error) {
	var ID pgtype.UUID
	if err := ID.Scan(id); err != nil {
		return ID, fmt.Errorf("invalid uuid: %w", err)
	}

	return ID, nil
}

func UUIDString(id pgtype.UUID) (string, error) {
	if !id.Valid {
		return "", fmt.Errorf("invalid uuid")
	}
	return uuid.UUID(id.Bytes).String(), nil
}
