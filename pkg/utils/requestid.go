package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRequestID creates a unique identifier for each request
func GenerateRequestID() (string, error) {
	bytes := make([]byte, 8) // 8 bytes = 16 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate request ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
