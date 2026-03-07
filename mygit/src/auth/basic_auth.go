package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
)

var (
	adminUsername string
	adminPassword string
)

func init() {
	adminUsername = getEnvDefault("ADMIN_USERNAME", "admin")
	adminPassword = getEnvDefault("ADMIN_PASSWORD", "admin")
}

func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey := r.Header.Get("Authorization"); apiKey != "" {
			if strings.HasPrefix(apiKey, "Bearer ") {
				key := strings.TrimPrefix(apiKey, "Bearer ")
				if valid, _ := ValidateAPIKey(key); valid {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			setWWWAuthenticate(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !validateCredentials(username, password) {
			setWWWAuthenticate(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(withUserContext(r.Context(), username))
		next.ServeHTTP(w, r)
	})
}

func validateCredentials(username, password string) bool {
	return username == adminUsername && password == adminPassword
}

func setWWWAuthenticate(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="MyGit"`)
}

func withUserContext(ctx context.Context, username string) context.Context {
	return ctx
}

func getEnvDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

var ErrUnauthorized = errors.New("unauthorized")
