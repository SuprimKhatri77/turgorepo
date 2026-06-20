package utils

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func ConvertToUUID(id string) (pgtype.UUID, error) {
	var ID pgtype.UUID
	if err := ID.Scan(id); err != nil {
		return ID, fmt.Errorf("invalid uuid: %w", err)
	}

	return ID, nil
}
