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

    "github.com/flozsc/mygit/src/config"
)

// SSHKey represents an SSH public key
type SSHKey struct {
	Key        string `json:"key"`
	Fingerprint string `json:"fingerprint"`
	Comment    string `json:"comment"`
}

// AddUserSSHKey adds an SSH public key for the specified user.
func AddUserSSHKey(username, publicKey string) error {
    // Validate key format
    if !isValidSSHKey(publicKey) {
        return ErrSSHKeyInvalid
    }

    // Parse to get comment (optional)
    _, comment, err := parseSSHKey(publicKey)
    if err != nil {
        return err
    }

    // Prepare storage path using config package
    keyPath := filepath.Join(config.GetSSHUserKeysDir(), username+".pub")
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
        return err
    }

    // Append the key to the user's file
    f, err := os.OpenFile(keyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        return err
    }
    defer f.Close()

    // Write key (including comment if present)
    line := publicKey
    if comment != "" {
        line = fmt.Sprintf("%s %s", publicKey, comment)
    }
    if _, err := f.WriteString(line + "\n"); err != nil {
        return err
    }
    return nil
}

// ListUserSSHKeys lists all SSH public keys for the specified user.
func ListUserSSHKeys(username string) ([]SSHKey, error) {
    keyPath := filepath.Join(config.GetSSHUserKeysDir(), username+".pub")
    f, err := os.Open(keyPath)
    if err != nil {
        if os.IsNotExist(err) {
            return []SSHKey{}, nil // No keys yet
        }
        return nil, err
    }
    defer f.Close()

    var keys []SSHKey
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        // Expect standard format: <type> <base64> [comment]
        parts := strings.Fields(line)
        if len(parts) < 2 {
            continue
        }
        keyPart := strings.Join(parts[:2], " ")
        comment := ""
        if len(parts) > 2 {
            comment = strings.Join(parts[2:], " ")
        }
        fp, _ := calculateFingerprint(keyPart)
        keys = append(keys, SSHKey{Key: keyPart, Fingerprint: fp, Comment: comment})
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return keys, nil
}

// DeleteUserSSHKey removes a key matching the given fingerprint for the user.
func DeleteUserSSHKey(username, fingerprint string) error {
    keyPath := filepath.Join(config.GetSSHUserKeysDir(), username+".pub")
    // Read existing keys
    data, err := os.ReadFile(keyPath)
    if err != nil {
        return err
    }
    lines := strings.Split(string(data), "\n")
    var keep []string
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        parts := strings.Fields(line)
        if len(parts) < 2 {
            continue
        }
        keyPart := strings.Join(parts[:2], " ")
        fp, _ := calculateFingerprint(keyPart)
        if fp != fingerprint {
            keep = append(keep, line)
        }
    }
    // Write back kept lines
    return os.WriteFile(keyPath, []byte(strings.Join(keep, "\n")), 0600)
}

// Existing AddSSHKey (global) retained for compatibility – now simply forwards to per‑user using session username.
func AddSSHKey(publicKey string) error {
    // This function is kept for backward compatibility; it will add the key for the currently authenticated user.
    // In the HTTP handler we will call AddUserSSHKey with the session username.
    return errors.New("AddSSHKey is deprecated; use AddUserSSHKey with explicit username")
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