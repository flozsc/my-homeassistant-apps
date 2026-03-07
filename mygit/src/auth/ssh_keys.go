package auth

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSHKey represents an SSH public key
type SSHKey struct {
	Key        string `json:"key"`
	Fingerprint string `json:"fingerprint"`
	Comment    string `json:"comment"`
}

// AddSSHKey adds an SSH public key for a user
func AddSSHKey(publicKey string) error {
	// Validate key format
	if !isValidSSHKey(publicKey) {
		return errors.New("invalid SSH key format")
	}

	// Parse key to extract components
	_, comment, err := parseSSHKey(publicKey)
	if err != nil {
		return err
	}
	
	// Calculate fingerprint (for future use)
	_, err = calculateFingerprint(publicKey)
	if err != nil {
		return err
	}
	
	// Store in authorized_keys format
	authKey := fmt.Sprintf("%s %s", publicKey, comment)
	
	// Append to authorized_keys file
	return appendToAuthorizedKeys(authKey)
}

// isValidSSHKey checks if a string is a valid SSH public key
func isValidSSHKey(key string) bool {
	// Basic format check
	parts := strings.Fields(key)
	if len(parts) < 2 {
		return false
	}
	
	// Check for common key types
	validTypes := []string{"ssh-rsa", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521", "ssh-ed25519"}
	for _, prefix := range validTypes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	
	return false
}

// parseSSHKey parses an SSH key into components
func parseSSHKey(key string) (string, string, error) {
	// Split by spaces
	parts := strings.Fields(key)
	if len(parts) < 2 {
		return "", "", errors.New("invalid key format")
	}
	
	// Last part is the comment
	comment := parts[len(parts)-1]
	// Everything except last part is the key
	keyPart := strings.Join(parts[:len(parts)-1], " ")
	
	return keyPart, comment, nil
}

// calculateFingerprint calculates SSH key fingerprint
func calculateFingerprint(key string) (string, error) {
	// TODO: Implement actual fingerprint calculation
	// For now, use a simple hash of the key
	h := sha256.New()
	h.Write([]byte(key))
	fingerprint := hex.EncodeToString(h.Sum(nil))
	return fingerprint, nil
}

// appendToAuthorizedKeys appends a key to the authorized_keys file
func appendToAuthorizedKeys(key string) error {
	authFile := filepath.Join("/home/git", ".ssh", "authorized_keys")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(authFile), 0700); err != nil {
		return err
	}
	
	// Open file for appending
	f, err := os.OpenFile(authFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	
	// Write the key
	writer := bufio.NewWriter(f)
	if _, err := writer.WriteString(key + "\n"); err != nil {
		return err
	}
	
	return writer.Flush()
}

// SSHKeyError represents SSH key errors
var (
	ErrSSHKeyInvalid    = errors.New("invalid SSH key format")
	ErrSSHKeyExists     = errors.New("SSH key already exists")
	ErrSSHKeyWriteFailed = errors.New("failed to write SSH key")
)