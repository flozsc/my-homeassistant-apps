package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// BasicAuthMiddleware provides Basic Authentication middleware
func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API key first
		if apiKey := r.Header.Get("Authorization"); apiKey != "" {
			if strings.HasPrefix(apiKey, "Bearer ") {
				key := strings.TrimPrefix(apiKey, "Bearer ")
				if valid, _ := ValidateAPIKey(key); valid {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// Check for Basic Auth
		username, password, ok := r.BasicAuth()
		if !ok {
			setWWWAuthenticate(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate credentials
		if !validateCredentials(username, password) {
			setWWWAuthenticate(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		r = r.WithContext(withUserContext(r.Context(), username))

		next.ServeHTTP(w, r)
	})
}

// validateCredentials checks username and password
func validateCredentials(username, password string) bool {
	// TODO: Implement actual credential validation
	// For development, accept admin/admin
	return username == "admin" && password == "admin"
}

// setWWWAuthenticate sets the WWW-Authenticate header
func setWWWAuthenticate(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="MyGit"`)
}

// withUserContext adds user information to the context
func withUserContext(ctx context.Context, username string) context.Context {
	// TODO: Implement context with user info
	return ctx
}

// AuthError represents authentication errors
var ErrUnauthorized = errors.New("unauthorized")