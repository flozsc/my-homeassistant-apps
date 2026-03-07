package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	Version = "1.0.0"
	AppName = "mygit"
)

func main() {
	// Configuration
	config := struct {
		HTTPPort    string
		RepoStorage string
		AdminUser   string
		AdminPass   string
	}{
		HTTPPort:    getEnv("HTTP_PORT", "3000"),
		RepoStorage: getEnv("REPO_STORAGE", "/data/repos"),
		AdminUser:   getEnv("ADMIN_USERNAME", "admin"),
		AdminPass:   getEnv("ADMIN_PASSWORD", ""), // Will prompt if empty
	}

	// Ensure repository storage exists
	if err := os.MkdirAll(config.RepoStorage, 0755); err != nil {
		log.Fatalf("Failed to create repo storage: %v", err)
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to %s v%s!", AppName, Version)
	})

	// Start server
	addr := ":" + config.HTTPPort
	log.Printf("Starting %s v%s on %s", AppName, Version, addr)
	log.Printf("Repository storage: %s", config.RepoStorage)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// getEnv returns environment variable or default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}