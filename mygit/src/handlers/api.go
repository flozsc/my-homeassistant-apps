package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flozsc/mygit/src/auth"
	"github.com/go-chi/chi/v5"
)

type APIHandler struct{}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

func (h *APIHandler) RegisterRoutes(r chi.Router) {
	// Health check (no auth required)
	r.Get("/health", handleHealth)

	// Auth endpoints
	r.Post("/auth/login", auth.HandleLogin)
	r.Post("/auth/logout", auth.HandleLogout)
	r.Get("/auth/me", auth.HandleMe)

	// Repo endpoints
	r.Get("/repos", auth.HandleListRepos)
	r.Post("/repos", auth.HandleCreateRepo)
	r.Delete("/repos/{name}", auth.HandleDeleteRepo)

	// Repo detail routes - use a subrouter
	r.Route("/repos/{name}", func(r chi.Router) {
		r.Get("/", auth.HandleGetRepo)
		r.Get("/commits", auth.HandleGetCommits)
		r.Get("/branches", auth.HandleGetBranches)
		r.Get("/contributors", auth.HandleGetContributors)
		r.Get("/tree", auth.HandleGetTree)
		r.Get("/tree/{path:.*}", auth.HandleGetTree)
		r.Get("/blob/{path:.*}", auth.HandleGetBlob)

		// User management (admin only)
		r.Get("/users", auth.HandleListUsers)
		r.Post("/users", auth.HandleCreateUser)
		r.Delete("/users/{username}", auth.HandleDeleteUser)
	})

	// SSH key management (per‑user, not per‑repo)
	r.Get("/ssh-keys", auth.HandleListUserSSHKeys)
	r.Post("/ssh-keys", auth.HandleAddUserSSHKey)
	r.Delete("/ssh-keys/{fingerprint}", auth.HandleDeleteUserSSHKey)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "0.2.0",
	})
}
