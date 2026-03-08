package auth

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
)

// HandleListUserSSHKeys lists SSH keys for the authenticated user.
func HandleListUserSSHKeys(w http.ResponseWriter, r *http.Request) {
    sess, err := GetSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
        return
    }
    if !hasScope(sess.Scopes, "read") && !hasScope(sess.Scopes, "write") && !hasScope(sess.Scopes, "admin") {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Forbidden"})
        return
    }
    keys, err := ListUserSSHKeys(sess.Username)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(keys)
}

// HandleAddUserSSHKey adds a new SSH key for the authenticated user.
func HandleAddUserSSHKey(w http.ResponseWriter, r *http.Request) {
    sess, err := GetSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
        return
    }
    if !hasScope(sess.Scopes, "write") && !hasScope(sess.Scopes, "admin") {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Forbidden"})
        return
    }
    var req struct {
        Key string `json:"key"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
        return
    }
    if err := AddUserSSHKey(sess.Username, req.Key); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Key added"})
}

// HandleDeleteUserSSHKey deletes a key identified by its fingerprint for the authenticated user.
func HandleDeleteUserSSHKey(w http.ResponseWriter, r *http.Request) {
    sess, err := GetSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
        return
    }
    if !hasScope(sess.Scopes, "write") && !hasScope(sess.Scopes, "admin") {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Forbidden"})
        return
    }
    fingerprint := chi.URLParam(r, "fingerprint")
    if fingerprint == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Fingerprint required"})
        return
    }
    if err := DeleteUserSSHKey(sess.Username, fingerprint); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Key deleted"})
}
