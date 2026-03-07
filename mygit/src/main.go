package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"github.com/flozsc/mygit/src/auth"
)

const (
	Version = "0.0.1"
	AppName = "mygit"
)

var repoStorage string

func main() {
	// Configuration
	repoStorage = getEnv("REPO_STORAGE", "/data/repos")
	httpPort := getEnv("HTTP_PORT", "3000")

	// Ensure repository directory exists
	if err := os.MkdirAll(repoStorage, 0755); err != nil {
		log.Printf("Warning: Could not create repo storage (may already exist or permissions issue): %v", err)
		// Continue anyway - directory might exist or we're in a container with volume mounts
	}

	// Set up HTTP server
	mux := http.NewServeMux()

	// Web routes
	mux.HandleFunc("/repos", handleRepoList)
	mux.HandleFunc("/repos/", handleRepo)

	// Main handler that routes both web and Git requests
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGitRequest(r) {
			handleGitSmartHTTP(w, r)
		} else {
			handleIndex(w, r)
		}
	})

	// Wrap with authentication middleware
	wrappedHandler := auth.BasicAuthMiddleware(mainHandler)
	mux.Handle("/", wrappedHandler)

	// Start server
	addr := ":" + httpPort
	log.Printf("Starting %s v%s on %s", AppName, Version, addr)
	log.Printf("Repository storage: %s", repoStorage)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// handleIndex serves the main page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	// For now, keep simple response
	// TODO: Replace with template rendering when we have a proper templating system
	fmt.Fprintf(w, "Welcome to %s v%s!", AppName, Version)
}

// handleRepoList lists all repositories
func handleRepoList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// List repositories
	repos, err := filepath.Glob(filepath.Join(repoStorage, "*.git"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"repos\": %d}", len(repos))
}

// handleRepo serves individual repository requests
func handleRepo(w http.ResponseWriter, r *http.Request) {
	repoName := strings.TrimPrefix(r.URL.Path, "/repos/")
	if repoName == "" || !strings.HasSuffix(repoName, ".git") {
		http.NotFound(w, r)
		return
	}
	
	repoPath := filepath.Join(repoStorage, repoName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	
	fmt.Fprintf(w, "Repository: %s", repoName)
}

// handleGitSmartHTTP handles Git smart HTTP protocol
func handleGitSmartHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this is a Git request
	if !isGitRequest(r) {
		return
	}
	
	// Extract repository name
	repoName := extractRepoName(r.URL.Path)
	if repoName == "" {
		http.NotFound(w, r)
		return
	}
	
	repoPath := filepath.Join(repoStorage, repoName)
	
	// Check if repository exists, create if not (auto-create on first push)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if err := createBareRepo(repoPath); err != nil {
				http.Error(w, fmt.Sprintf("Failed to create repository: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
		http.NotFound(w, r)
			return
		}
	}
	
	// Handle Git smart HTTP
	service := r.URL.Query().Get("service")
	if service == "" {
		// Advertise services
		advertiseGitServices(w)
		return
	}
	
	switch service {
	case "git-upload-pack":
		handleUploadPack(w, r, repoPath)
	case "git-receive-pack":
		handleReceivePack(w, r, repoPath)
	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

// isGitRequest checks if the request is for Git operations
func isGitRequest(r *http.Request) bool {
	// Check for Git-specific headers or query parameters
	if r.Header.Get("Git-Protocol") != "" {
		return true
	}
	if r.URL.Query().Get("service") != "" {
		return true
	}
	// Check path pattern
	if strings.HasSuffix(r.URL.Path, ".git/info/refs") {
		return true
	}
	if strings.Contains(r.URL.Path, ".git/git-") {
		return true
	}
	return false
}

// extractRepoName extracts repository name from URL path
func extractRepoName(path string) string {
	// Remove .git suffix if present
	name := strings.TrimSuffix(path, ".git")
	// Get last component
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Validate name
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		return ""
	}
	return name + ".git"
}

// createBareRepo creates a new bare Git repository
func createBareRepo(repoPath string) error {
	cmd := exec.Command("git", "init", "--bare", repoPath)
	return cmd.Run()
}

// advertiseGitServices advertises available Git services
func advertiseGitServices(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	fmt.Fprintf(w, "001e# service=git-upload-pack\n0000001f# service=git-receive-pack\n0000")
}

// handleUploadPack handles git-upload-pack service
func handleUploadPack(w http.ResponseWriter, r *http.Request, repoPath string) {
	// Set required headers
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")
	
	// Execute git-upload-pack
	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Env = append(os.Environ(), "GIT_HTTP_EXPORT_ALL=")
	
	if err := cmd.Run(); err != nil {
		log.Printf("upload-pack failed: %v", err)
	}
}

// handleReceivePack handles git-receive-pack service
func handleReceivePack(w http.ResponseWriter, r *http.Request, repoPath string) {
	// Set required headers
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	w.Header().Set("Cache-Control", "no-cache")
	
	// Execute git-receive-pack
	cmd := exec.Command("git", "receive-pack", "--stateless-rpc", repoPath)
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Env = append(os.Environ(), "GIT_HTTP_EXPORT_ALL=")
	
	if err := cmd.Run(); err != nil {
		log.Printf("receive-pack failed: %v", err)
	}
}

// getEnv returns environment variable or default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}