package config

import (
    "crypto/rand"
    "encoding/base64"
    "os"
)

var (
    sshUserKeysDir string
    tokenSecret    string
)

func init() {
    // SSH user keys directory, default to /data/ssh/users
    if dir := os.Getenv("SSH_USER_KEYS_DIR"); dir != "" {
        sshUserKeysDir = dir
    } else {
        sshUserKeysDir = "/data/ssh/users"
    }

    // Token secret – generate random 32‑byte base64 if not provided
    if sec := os.Getenv("TOKEN_SECRET"); sec != "" {
        tokenSecret = sec
    } else {
        b := make([]byte, 32)
        _, _ = rand.Read(b)
        tokenSecret = base64.URLEncoding.EncodeToString(b)
    }
}

func GetSSHUserKeysDir() string { return sshUserKeysDir }
func GetTokenSecret() string   { return tokenSecret }
