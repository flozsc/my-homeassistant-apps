package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/flozsc/mygit/src/git"
	"github.com/go-chi/chi/v5"
)

var (
	sessions = &SessionStore{
		data: make(map[string]*Session),
	}
	validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

type Session struct {
	Username  string
	Token     string
	ExpiresAt time.Time
	Scopes    []string
}

type SessionStore struct {
	mu   sync.RWMutex
	data map[string]*Session
}

func Init(username, password string) {
	adminUsername = username
	adminPassword = password
	git.SetStorage("/data/repos")
}

func SetRepoStorage(path string) {
	git.SetStorage(path)
}

func (s *SessionStore) Create(username string) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Retrieve user scopes
	scopes, err := GetUserStore().GetScopes(username)
	if err != nil {
		return nil, err
	}

	session := &Session{
		Username:  username,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scopes:    scopes,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[token] = session

	return session, nil
}

func (s *SessionStore) Get(token string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.data[token]
	if !ok {
		return nil, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, token)
}

func (s *SessionStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, session := range s.data {
		if now.After(session.ExpiresAt) {
			delete(s.data, token)
		}
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func ValidateCredentials(username, password string) bool {
	return username == adminUsername && password == adminPassword
}

// API handlers

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	if !ValidateCredentials(req.Username, req.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid credentials"})
		return
	}

	session, err := sessions.Create(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Failed to create session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token:    session.Token,
		Username: session.Username,
	})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := GetTokenFromRequest(r)
	if token != "" {
		sessions.Delete(token)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}

type UserInfo struct {
	Username string `json:"username"`
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := GetSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserInfo{
		Username: session.Username,
	})
}

func GetTokenFromRequest(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
	}

	// Check cookie
	cookie, err := r.Cookie("token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func GetSession(r *http.Request) (*Session, error) {
	token := GetTokenFromRequest(r)
	if token == "" {
		return nil, errors.New("no token")
	}
	return sessions.Get(token)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health check without auth
		if r.URL.Path == "/api/v1/health" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key (from existing auth)
		if apiKey := r.Header.Get("Authorization"); apiKey != "" {
			if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
				token := apiKey[7:]
				if _, err := sessions.Get(token); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), "username", "admin"))
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// Check Basic auth
		username, password, ok := r.BasicAuth()
		if ok && ValidateCredentials(username, password) {
			r = r.WithContext(context.WithValue(r.Context(), "username", username))
			next.ServeHTTP(w, r)
			return
		}

		// No valid auth
		w.Header().Set("WWW-Authenticate", `Basic realm="MyGit"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// Repo handlers

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RepoInfo struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
	LastCommit    string `json:"last_commit,omitempty"`
	Size          string `json:"size,omitempty"`
	URL           string `json:"url"`
}

func HandleCreateRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name is required"})
		return
	}

	if !validName.MatchString(req.Name) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid repository name"})
		return
	}

	if err := git.CreateBareRepo(req.Name); err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	defaultBranch, _ := git.GetDefaultBranch(req.Name)
	lastCommit, _ := git.GetLatestCommit(req.Name, "HEAD")
	size, _ := git.GetRepoSize(req.Name)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RepoInfo{
		Name:          req.Name,
		Description:   req.Description,
		DefaultBranch: defaultBranch,
		LastCommit:    lastCommit,
		Size:          size,
		URL:           "/api/v1/repos/" + req.Name,
	})
}

func HandleDeleteRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := extractParam(r, "name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name required"})
		return
	}

	if err := git.DeleteRepo(name); err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Repository deleted"})
}

func HandleListRepos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	storage := git.GetStoragePath()
	entries, err := getRepoList(storage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Failed to list repositories: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

func getRepoList(storage string) ([]RepoInfo, error) {
	entries, err := os.ReadDir(storage)
	if err != nil {
		return nil, err
	}

	var repos []RepoInfo
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// Check if it's a .git directory
			if !strings.HasSuffix(name, ".git") {
				name = name + ".git"
			}
			defaultBranch, _ := git.GetDefaultBranch(strings.TrimSuffix(name, ".git"))
			lastCommit, _ := git.GetLatestCommit(strings.TrimSuffix(name, ".git"), "HEAD")
			size, _ := git.GetRepoSize(strings.TrimSuffix(name, ".git"))

			repos = append(repos, RepoInfo{
				Name:          strings.TrimSuffix(name, ".git"),
				DefaultBranch: defaultBranch,
				LastCommit:    lastCommit,
				Size:          size,
				URL:           "/api/v1/repos/" + strings.TrimSuffix(name, ".git"),
			})
		}
	}
	return repos, nil
}

func HandleGetRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := extractParam(r, "name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name required"})
		return
	}

	if !git.RepoExists(name) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository not found"})
		return
	}

	defaultBranch, _ := git.GetDefaultBranch(name)
	lastCommit, _ := git.GetLatestCommit(name, "HEAD")
	size, _ := git.GetRepoSize(name)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RepoInfo{
		Name:          name,
		DefaultBranch: defaultBranch,
		LastCommit:    lastCommit,
		Size:          size,
		URL:           "/api/v1/repos/" + name,
	})
}

func HandleGetCommits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := extractParam(r, "name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name required"})
		return
	}

	if !git.RepoExists(name) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository not found"})
		return
	}

	branch := r.URL.Query().Get("branch")
	if branch == "" {
		branch, _ = git.GetDefaultBranch(name)
	}

	limit := 30
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	commits, err := git.GetCommits(r.Context(), name, branch, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(commits)
}

func HandleGetBranches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := extractParam(r, "name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name required"})
		return
	}

	if !git.RepoExists(name) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository not found"})
		return
	}

	branches, err := git.GetBranches(r.Context(), name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(branches)
}

func HandleGetTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := extractParam(r, "name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name required"})
		return
	}

	if !git.RepoExists(name) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository not found"})
		return
	}

	// Get ref from query param first
	ref := r.URL.Query().Get("ref")
	path := chi.URLParam(r, "path")

	// If no query param ref and path looks like a branch, use it as ref
	if ref == "" && path != "" && !strings.Contains(path, "/") {
		ref = path
		path = ""
	}

	// If still no ref, get default branch
	if ref == "" {
		ref, _ = git.GetDefaultBranch(name)
	}

	entries, err := git.GetTree(r.Context(), name, ref, path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

func HandleGetBlob(w http.ResponseWriter, r *http.Request) {
	name := extractParam(r, "name")
	if name == "" {
		http.Error(w, "Repository name required", http.StatusBadRequest)
		return
	}

	path := chi.URLParam(r, "path")
	if path == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	ref := r.URL.Query().Get("ref")
	// If path contains a slash, the first part might be the ref
	if ref == "" && strings.Contains(path, "/") {
		parts := strings.SplitN(path, "/", 2)
		ref = parts[0]
		path = parts[1]
	}
	if ref == "" {
		ref, _ = git.GetDefaultBranch(name)
	}

	content, err := git.GetBlob(name, ref, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func HandleGetContributors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := extractParam(r, "name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository name required"})
		return
	}

	if !git.RepoExists(name) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Repository not found"})
		return
	}

	contributors, err := git.GetContributors(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contributors)
}

func extractParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func extractPathParam(r *http.Request, name string) string {
	return chi.URLParam(r, "path")
}

func StartSessionCleanup() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sessions.Cleanup()
		}
	}()
}
