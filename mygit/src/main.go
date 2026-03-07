package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flozsc/mygit/src/auth"
)

const (
	Version = "0.1.4"
	AppName = "mygit"
)

var (
	repoStorage  string
	templates    *template.Template
	staticDir    string
)

type PageData struct {
	Title         string
	Version       string
	Authenticated bool
	Username      string
	Repos         []Repo
	ErrorMessage  string
	Repo          *Repo
}

type Repo struct {
	Name        string
	Description string
	LastCommit  string
	Size        string
	URL         string
}

func main() {
	repoStorage = getEnv("REPO_STORAGE", "/data/repos")
	httpPort := getEnv("HTTP_PORT", "3000")
	staticDir = "/data/static"

	if err := os.MkdirAll(repoStorage, 0755); err != nil {
		log.Printf("Warning: Could not create repo storage: %v", err)
	}

	if err := os.MkdirAll(staticDir, 0755); err != nil {
		log.Printf("Warning: Could not create static dir: %v", err)
	}

	var err error
	templates, err = template.ParseGlob("/data/web/templates/*.html")
	if err != nil {
		log.Printf("Warning: Could not parse templates: %v", err)
		templates = template.New("fallback")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/repos", handleRepoList)
	mux.HandleFunc("/repos/", handleRepo)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/data/web/static"))))
	mux.Handle("/favicon.ico", http.FileServer(http.Dir("/data/web/static")))

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGitRequest(r) {
			handleGitSmartHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	})

	wrappedHandler := auth.BasicAuthMiddleware(mainHandler)
	mux.Handle("/", wrappedHandler)

	addr := ":" + httpPort
	log.Printf("Starting %s v%s on %s", AppName, Version, addr)
	log.Printf("Repository storage: %s", repoStorage)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title:   "Home",
		Version: Version,
	}

	if renderTemplate(w, "base.html", "index.html", data) != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleRepoList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repos, err := filepath.Glob(filepath.Join(repoStorage, "*.git"))
	if err != nil {
		data := PageData{
			Title:        "Repositories",
			Version:      Version,
			ErrorMessage: "Failed to list repositories",
		}
		renderTemplate(w, "base.html", "repos.html", data)
		return
	}

	var repoList []Repo
	for _, r := range repos {
		name := filepath.Base(r)
		repoList = append(repoList, Repo{
			Name: strings.TrimSuffix(name, ".git"),
			URL:  "/repos/" + name,
		})
	}

	data := PageData{
		Title: "Repositories",
		Repos: repoList,
	}

	if renderTemplate(w, "base.html", "repos.html", data) != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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

	data := PageData{
		Title: repoName,
		Repo: &Repo{
			Name: strings.TrimSuffix(repoName, ".git"),
			URL:  "/" + repoName,
		},
	}

	renderTemplate(w, "base.html", "repos.html", data)
}

func renderTemplate(w http.ResponseWriter, base, content string, data PageData) error {
	data.Version = Version
	return templates.ExecuteTemplate(w, base, data)
}

func handleGitSmartHTTP(w http.ResponseWriter, r *http.Request) {
	if !isGitRequest(r) {
		return
	}

	repoName := extractRepoName(r.URL.Path)
	if repoName == "" {
		http.NotFound(w, r)
		return
	}

	repoPath := filepath.Join(repoStorage, repoName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if err := createBareRepo(repoPath); err != nil {
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

	switch service {
	case "git-upload-pack":
		handleUploadPack(w, r, repoPath)
	case "git-receive-pack":
		handleReceivePack(w, r, repoPath)
	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

func isGitRequest(r *http.Request) bool {
	if r.Header.Get("Git-Protocol") != "" {
		return true
	}
	if r.URL.Query().Get("service") != "" {
		return true
	}
	if strings.HasSuffix(r.URL.Path, ".git/info/refs") {
		return true
	}
	if strings.Contains(r.URL.Path, ".git/git-") {
		return true
	}
	return false
}

func extractRepoName(path string) string {
	name := strings.TrimSuffix(path, ".git")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
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
	cmd.Stderr = w
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
	cmd.Stderr = w
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
