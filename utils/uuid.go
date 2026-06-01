package utils

import "github.com/google/uuid"

// NewUUID returns a new random (v4) UUID string.
func NewUUID() string { return uuid.New().String() }

// IsValidUUID reports whether s is a valid UUID.
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
