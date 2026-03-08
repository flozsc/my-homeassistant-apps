package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flozsc/mygit/src/auth"
	"github.com/flozsc/mygit/src/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed ui/*
var uiAssets embed.FS

const (
	Version = "0.2.0"
	AppName = "mygit"
)

var (
	repoStorage string
	staticDir   string
)

func main() {
	repoStorage = getEnv("REPO_STORAGE", "/data/repos")
	httpPort := getEnv("HTTP_PORT", "3000")
	adminUsername := getEnv("ADMIN_USERNAME", "admin")
	adminPassword := getEnv("ADMIN_PASSWORD", "admin")

	staticDir = getEnv("STATIC_DIR", "/data/static")

	if err := os.MkdirAll(repoStorage, 0755); err != nil {
		log.Printf("Warning: Could not create repo storage: %v", err)
	}

	// Initialize auth
	// Load users from data/users.json and ensure admin exists
	if err := auth.GetUserStore().Load("./data/users.json"); err != nil {
		log.Fatalf("Failed to load users: %v", err)
	}
	// Ensure admin user exists (create if missing)
	if valid, _, err := auth.GetUserStore().ValidatePassword(adminUsername, adminPassword); err != nil || !valid {
		// Create admin user with admin scope if not present
		if _, err := auth.GetUserStore().Create(adminUsername, adminPassword, []string{"admin"}); err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
	}
	auth.SetRepoStorage(repoStorage)

	// Start session cleanup
	auth.StartSessionCleanup()

	// Create chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	apiHandler := handlers.NewAPIHandler()
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		apiHandler.RegisterRoutes(r)
	})

	// Git smart protocol - must be before catch-all
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		if isGitRequest(r.URL.Path) {
			handleGitSmartHTTP(w, r)
		} else {
			serveUI(w, r)
		}
	})

	addr := ":" + httpPort
	log.Printf("Starting %s v%s on %s", AppName, Version, addr)
	log.Printf("Repository storage: %s", repoStorage)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Don't serve UI for API or git paths
	if strings.HasPrefix(path, "/api/") || isGitRequest(path) {
		http.NotFound(w, r)
		return
	}

	// Serve static assets from embedded UI
	if path == "/" || path == "" {
		index, err := uiAssets.ReadFile("ui/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
		return
	}

	// Try to serve from embedded UI
	staticPath := strings.TrimPrefix(path, "/")
	data, err := uiAssets.ReadFile(staticPath)
	if err != nil {
		// Fallback to serving index.html for SPA routes
		if !strings.Contains(path, ".") {
			index, err := uiAssets.ReadFile("ui/index.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(index)
			return
		}
		http.NotFound(w, r)
		return
	}

	// Determine content type
	contentType := getContentType(staticPath)
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	default:
		return "text/plain; charset=utf-8"
	}
}

func isGitRequest(path string) bool {
	if strings.HasSuffix(path, ".git/info/refs") {
		return true
	}
	if strings.Contains(path, ".git/git-") {
		return true
	}
	if strings.Contains(path, "/git-upload-pack") || strings.Contains(path, "/git-receive-pack") {
		return true
	}
	return false
}

func handleGitSmartHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	repoName := extractRepoName(path)
	if repoName == "" {
		http.NotFound(w, r)
		return
	}

	repoPath := filepath.Join(repoStorage, repoName)

	// Auto-create repo on push if it doesn't exist
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if err := createBareRepo(repoPath); err != nil {
				log.Printf("Failed to create repo: %v", err)
				http.Error(w, "Failed to create repository", http.StatusInternalServerError)
				return
			}
		} else {
			http.NotFound(w, r)
			return
		}
	}

	service := r.URL.Query().Get("service")
	if service == "" {
		advertiseGitServices(w)
		return
	}

	// Enforce scope for git operations
	if err := auth.CheckGitScope(r, service); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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

func extractRepoName(path string) string {
	// Handle paths like /some-repo.git/info/refs or /some-repo/info/refs
	name := strings.TrimSuffix(path, ".git/info/refs")
	name = strings.TrimSuffix(name, "/info/refs")
	name = strings.TrimSuffix(name, "/git-upload-pack")
	name = strings.TrimSuffix(name, "/git-receive-pack")

	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}

	// Remove .git suffix if present
	name = strings.TrimSuffix(name, ".git")

	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") {
		return ""
	}
	return name + ".git"
}

func createBareRepo(repoPath string) error {
	cmd := exec.Command("git", "init", "--bare", repoPath)
	return cmd.Run()
}

func advertiseGitServices(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	w.Write([]byte("001e# service=git-upload-pack\n0000001f# service=git-receive-pack\n0000"))
}

func handleUploadPack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GIT_HTTP_EXPORT_ALL=")

	if err := cmd.Run(); err != nil {
		log.Printf("upload-pack failed: %v", err)
	}
}

func handleReceivePack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	cmd := exec.Command("git", "receive-pack", "--stateless-rpc", repoPath)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GIT_HTTP_EXPORT_ALL=")

	if err := cmd.Run(); err != nil {
		log.Printf("receive-pack failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
