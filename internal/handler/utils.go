package handler

import (
	"fmt"

	"github.com/google/uuid"
)

// parseUUID parses a UUID string and returns the UUID
func parseUUID(uuidStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format: %s", uuidStr)
	}
	return id, nil
}

// validateUUID validates that a string is a valid UUID
func validateUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}