package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// APIKey represents an API key for authentication
type APIKey struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Scopes    []string  `json:"scopes"`
}

// GenerateAPIKey creates a new API key
func GenerateAPIKey(scopes []string) (*APIKey, error) {
	// Generate random key (32 bytes = 64 hex chars)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	now := time.Now()
	
	return &APIKey{
		ID:        generateID(),
		Key:       hex.EncodeToString(keyBytes),
		CreatedAt: now,
		ExpiresAt: now.AddDate(1, 0, 0), // 1 year from now
		Scopes:    scopes,
	}, nil
}

// ValidateAPIKey checks if an API key is valid
func ValidateAPIKey(key string) (bool, error) {
	// TODO: Implement actual validation against storage
	// For now, accept any non-empty key for development
	return key != "", nil
}

// generateID creates a random ID for API keys
func generateID() string {
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		// Fallback to timestamp if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(idBytes)
}

// APIKeyError represents API key related errors
var (
	ErrAPIKeyInvalid      = errors.New("invalid API key")
	ErrAPIKeyExpired      = errors.New("API key expired")
	ErrAPIKeyInsufficient = errors.New("insufficient permissions")
)