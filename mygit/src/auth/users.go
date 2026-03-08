package auth

import (
    "encoding/json"
    "errors"
    "os"
    "sync"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    Username string   `json:"username"`
    Password string   `json:"password_hash"`
    Scopes   []string `json:"scopes"`
}

type UserStore struct {
    mu   sync.RWMutex
    users map[string]*User
    path string
}

var (
    userStore = &UserStore{users: make(map[string]*User)}
)

func (us *UserStore) Load(path string) error {
    us.mu.Lock()
    defer us.mu.Unlock()
    us.path = path
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // initialize empty file
            us.users = make(map[string]*User)
            return nil
        }
        return err
    }
    var list []User
    if err := json.Unmarshal(data, &list); err != nil {
        return err
    }
    us.users = make(map[string]*User)
    for i := range list {
        u := list[i]
        us.users[u.Username] = &u
    }
    return nil
}

func (us *UserStore) Save() error {
    // Note: Save() does NOT acquire the mutex - the caller must hold it
    list := make([]User, 0, len(us.users))
    for _, u := range us.users {
        list = append(list, *u)
    }
    data, err := json.MarshalIndent(list, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(us.path, data, 0644)
}

func (us *UserStore) Create(username, password string, scopes []string) (*User, error) {
    us.mu.Lock()
    if _, exists := us.users[username]; exists {
        us.mu.Unlock()
        return nil, errors.New("user already exists")
    }
    if !validName.MatchString(username) {
        us.mu.Unlock()
        return nil, errors.New("invalid username")
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        us.mu.Unlock()
        return nil, err
    }
    u := &User{Username: username, Password: string(hash), Scopes: scopes}
    us.users[username] = u
    // Call Save without the lock since Save now expects the caller to hold it
    err = us.Save()
    us.mu.Unlock()
    return u, err
}

func (us *UserStore) ValidatePassword(username, password string) (bool, []string, error) {
    us.mu.RLock()
    defer us.mu.RUnlock()
    u, ok := us.users[username]
    if !ok {
        return false, nil, nil
    }
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    if err != nil {
        return false, nil, nil
    }
    return true, u.Scopes, nil
}

func (us *UserStore) GetScopes(username string) ([]string, error) {
    us.mu.RLock()
    defer us.mu.RUnlock()
    u, ok := us.users[username]
    if !ok {
        return nil, errors.New("user not found")
    }
    return u.Scopes, nil
}

func GetUserStore() *UserStore { return userStore }

// GetAll returns a slice of all users (without password hashes)
func (us *UserStore) GetAll() []User {
    us.mu.RLock()
    defer us.mu.RUnlock()
    list := make([]User, 0, len(us.users))
    for _, u := range us.users {
        list = append(list, User{Username: u.Username, Scopes: u.Scopes})
    }
    return list
}

// Delete removes a user from the store and persists the change.
func (us *UserStore) Delete(username string) error {
    us.mu.Lock()
    defer us.mu.Unlock()
    if _, ok := us.users[username]; !ok {
        return errors.New("user not found")
    }
    delete(us.users, username)
    return us.Save()
}

