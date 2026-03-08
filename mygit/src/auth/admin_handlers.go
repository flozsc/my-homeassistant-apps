package auth

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
)

type UserInfoResponse struct {
    Username string   `json:"username"`
    Scopes   []string `json:"scopes"`
}

type UsersListResponse struct {
    Users []UserInfoResponse `json:"users"`
}

// HandleListUsers returns all users (admin only)
func HandleListUsers(w http.ResponseWriter, r *http.Request) {
    sess, err := GetSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
        return
    }
    // Check admin scope
    if !hasScope(sess.Scopes, "admin") {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Forbidden"})
        return
    }
    users := GetUserStore().GetAll()
    resp := UsersListResponse{Users: []UserInfoResponse{}}
    for _, u := range users {
        resp.Users = append(resp.Users, UserInfoResponse{Username: u.Username, Scopes: u.Scopes})
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(resp)
}

// HandleCreateUser creates a new user (admin only)
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
    sess, err := GetSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
        return
    }
    if !hasScope(sess.Scopes, "admin") {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Forbidden"})
        return
    }
    var req struct {
        Username string   `json:"username"`
        Password string   `json:"password"`
        Scopes   []string `json:"scopes"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
        return
    }
    if req.Username == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Username and password required"})
        return
    }
    if _, err := GetUserStore().Create(req.Username, req.Password, req.Scopes); err != nil {
        w.WriteHeader(http.StatusConflict)
        json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(UserInfoResponse{Username: req.Username, Scopes: req.Scopes})
}

// HandleDeleteUser deletes a user (admin only)
func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
    sess, err := GetSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Not authenticated"})
        return
    }
    if !hasScope(sess.Scopes, "admin") {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Forbidden"})
        return
    }
    username := chi.URLParam(r, "username")
    if username == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Username required"})
        return
    }
    if err := GetUserStore().Delete(username); err != nil {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "User deleted"})
}

