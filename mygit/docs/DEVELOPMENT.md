# MyGit Development Guide

## Getting Started

### Prerequisites

- **Go 1.20+** - For building the application
- **Docker/Podman** - For container development
- **Git** - For version control
- **Make** - For build automation (optional)

### Project Setup

```bash
# Clone the repository
git clone https://github.com/flozsc/mygit
cd mygit

# Initialize Go module (if not already done)
go mod init github.com/flozsc/mygit

# Install dependencies
go mod tidy
```

## Development Workflow

### Local Development

```bash
# Build the application
go build -o mygit ./src/main.go

# Run the application
./mygit

# Access the web interface
open http://localhost:3000
```

### Container Development

```bash
# Build the Docker image
podman build -t mygit -f Dockerfile .

# Run the container
podman run -d \
  -p 3000:3000 \
  -p 2222:22 \
  -v $(pwd)/data:/data \
  mygit

# Access logs
podman logs -f mygit
```

### VS Code Dev Container

If using VS Code with the Remote-Containers extension:

1. Open the project in VS Code
2. Click "Reopen in Container" when prompted
3. The container will build with all dependencies
4. You can develop directly in the container environment

## Build System

### Makefile Targets

```makefile
# Build the application
build:
	go build -o mygit ./src/main.go

# Run tests
test:
	go test ./...

# Build Docker image
docker:
	podman build -t mygit -f Dockerfile .

# Run the application
run:
	./mygit

# Format code
format:
	gofmt -w .

# Lint code
lint:
	golint ./...
```

### Building for Different Architectures

```bash
# Build for ARM
GOARCH=arm GOARM=7 go build -o mygit_arm ./src/main.go

# Build for AMD64
GOARCH=amd64 go build -o mygit_amd64 ./src/main.go

# Cross-compile using Docker
podman run --rm -v $(pwd):/workspace -w /workspace golang:latest \
  GOARCH=arm64 go build -o mygit_arm64 ./src/main.go
```

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./... -v

# Run tests for specific package
go test ./src/git -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
```

### Integration Tests

```bash
# Start test server in background
./mygit &
SERVER_PID=$!

# Run integration tests
go test ./tests/integration -v

# Stop test server
kill $SERVER_PID
```

### End-to-End Tests

```bash
# Using the test script
./tests/e2e/run.sh

# Manual testing
curl http://localhost:3000
git clone http://localhost:3000/test-repo.git
```

## Code Structure

### Main Components

```
mygit/
├── src/
│   ├── main.go          # Entry point
│   ├── git/             # Git operations
│   │   ├── server.go    # Git server
│   │   └── repo.go      # Repository management
│   ├── web/             # Web interface
│   │   ├── server.go    # HTTP server
│   │   ├── routes.go    # Routing
│   │   └── handlers/    # Request handlers
│   ├── auth/            # Authentication
│   │   ├── api_keys.go  # API key management
│   │   ├── ssh_keys.go  # SSH key management
│   │   └── basic_auth.go # Basic authentication
│   └── config/          # Configuration
│       └── config.go    # Configuration loading
├── web/
│   ├── static/         # Static assets
│   └── templates/       # HTML templates
├── tests/
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── e2e/             # End-to-end tests
└── docs/                # Documentation
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `src/git` | Git protocol implementation |
| `src/web` | Web interface and API |
| `src/auth` | Authentication systems |
| `src/config` | Configuration management |

## Coding Standards

### Go Code

- Follow `gofmt` formatting
- Use `golint` for linting
- Write comprehensive docstrings
- Keep functions focused and short
- Handle errors explicitly

### Example Function

```go
// CreateRepository creates a new Git repository
// Returns the repository path or error
func CreateRepository(name string) (string, error) {
    // Validate repository name
    if !isValidRepoName(name) {
        return "", fmt.Errorf("invalid repository name: %s", name)
    }
    
    // Create repository path
    repoPath := filepath.Join(repoStorage, name+ ".git")
    
    // Initialize repository
    if err := os.MkdirAll(repoPath, 0755); err != nil {
        return "", fmt.Errorf("failed to create repo directory: %w", err)
    }
    
    // Run git init
    cmd := exec.Command("git", "init", "--bare", repoPath)
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("failed to initialize repo: %w", err)
    }
    
    return repoPath, nil
}
```

### HTML Templates

- Use Go's `html/template` package
- Keep templates simple and focused
- Use template inheritance where possible
- Escape all dynamic content

### CSS/JavaScript

- Use modern CSS (Flexbox, Grid)
- Minimal JavaScript
- Progressive enhancement
- Accessible markup

## Authentication Development

### API Keys

```go
// GenerateAPIKey creates a new API key
type APIKey struct {
    ID        string    `json:"id"`
    Key       string    `json:"key"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at,omitempty"`
    Scopes    []string  `json:"scopes"`
}

func GenerateAPIKey(userID string, scopes []string) (*APIKey, error) {
    // Generate random key
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        return nil, err
    }
    
    // Create API key structure
    apiKey := &APIKey{
        ID:        uuid.New().String(),
        Key:       hex.EncodeToString(key),
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year
        Scopes:    scopes,
    }
    
    // Store in database
    if err := db.CreateAPIKey(userID, apiKey); err != nil {
        return nil, err
    }
    
    return apiKey, nil
}
```

### SSH Keys

```go
// AddSSHKey adds a public key for a user
func AddSSHKey(userID string, publicKey string) error {
    // Validate key format
    if !isValidSSHKey(publicKey) {
        return fmt.Errorf("invalid SSH key format")
    }
    
    // Store in authorized_keys format
    authKey := fmt.Sprintf("%s %s", publicKey, userID)
    
    // Append to authorized_keys file
    if err := appendToAuthorizedKeys(authKey); err != nil {
        return err
    }
    
    // Store in database
    return db.AddSSHKey(userID, publicKey)
}
```

## Web Interface Development

### Routing

```go
func setupRoutes() *http.ServeMux {
    mux := http.NewServeMux()
    
    // Web routes
    mux.HandleFunc("/", handleIndex)
    mux.HandleFunc("/repos", handleRepoList)
    mux.HandleFunc("/repos/", handleRepo)
    mux.HandleFunc("/repos/new", handleRepoCreate)
    
    // API routes
    mux.HandleFunc("/api/repos", apiHandleRepos)
    mux.HandleFunc("/api/repos/", apiHandleRepo)
    
    // Auth routes
    mux.HandleFunc("/login", handleLogin)
    mux.HandleFunc("/logout", handleLogout)
    mux.HandleFunc("/api-keys", handleAPIKeys)
    mux.HandleFunc("/ssh-keys", handleSSHKeys)
    
    return mux
}
```

### Middleware

```go
// AuthMiddleware checks authentication
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check for API key
        if apiKey := r.Header.Get("Authorization"); apiKey != "" {
            if isValidAPIKey(apiKey) {
                next.ServeHTTP(w, r)
                return
            }
        }
        
        // Check for session cookie
        if cookie, err := r.Cookie("session"); err == nil {
            if isValidSession(cookie.Value) {
                next.ServeHTTP(w, r)
                return
            }
        }
        
        // Not authenticated
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
    })
}
```

## Git Server Development

### Smart HTTP Protocol

```go
// HandleGitSmartHTTP handles Git smart HTTP requests
func HandleGitSmartHTTP(w http.ResponseWriter, r *http.Request) {
    // Extract repository name from path
    repoName := extractRepoName(r.URL.Path)
    if repoName == "" {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    
    // Check authentication
    if !isAuthenticated(r) {
        w.Header().Set("WWW-Authenticate", `Basic realm="MyGit"`)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Handle Git smart HTTP
    repoPath := filepath.Join(repoStorage, repoName+ ".git")
    
    // Set up Git environment
    w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
    
    // Handle upload-pack or receive-pack
    service := r.URL.Query().Get("service")
    switch service {
    case "git-upload-pack":
        handleUploadPack(w, r, repoPath)
    case "git-receive-pack":
        handleReceivePack(w, r, repoPath)
    default:
        http.Error(w, "Bad Request", http.StatusBadRequest)
    }
}
```

## Database Development

### SQLite Schema

```sql
-- Users table
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- API keys table
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    scopes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- SSH keys table
CREATE TABLE ssh_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Repositories table
CREATE TABLE repositories (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    owner_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id)
);
```

### Database Access

```go
// DB represents the database connection
type DB struct {
    *sql.DB
}

// NewDB creates a new database connection
func NewDB(path string) (*DB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }
    
    // Set connection pool settings
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    
    // Initialize schema
    if err := initSchema(db); err != nil {
        return nil, err
    }
    
    return &DB{db}, nil
}

// initSchema initializes the database schema
func initSchema(db *sql.DB) error {
    schema := `
    CREATE TABLE IF NOT EXISTS users (...);
    CREATE TABLE IF NOT EXISTS api_keys (...);
    CREATE TABLE IF NOT EXISTS ssh_keys (...);
    CREATE TABLE IF NOT EXISTS repositories (...);
    `
    
    _, err := db.Exec(schema)
    return err
}
```

## Testing Strategy

### Test Organization

```
tests/
├── unit/
│   ├── git_test.go        # Git operations tests
│   ├── web_test.go        # Web handler tests
│   ├── auth_test.go       # Authentication tests
│   └── config_test.go     # Configuration tests
├── integration/
│   ├── api_test.go        # API integration tests
│   ├── git_integration.go # Git protocol tests
│   └── auth_integration.go # Auth flow tests
└── e2e/
    ├── setup.sh            # Test setup
    ├── test_basic.sh       # Basic functionality tests
    ├── test_auth.sh        # Authentication tests
    └── test_git.sh          # Git operations tests
```

### Test Examples

**Unit Test Example**
```go
func TestCreateRepository(t *testing.T) {
    // Setup
    tempDir := t.TempDir()
    repoStorage = tempDir
    
    // Test valid repository name
    repoPath, err := CreateRepository("test-repo")
    if err != nil {
        t.Fatalf("Failed to create repository: %v", err)
    }
    
    // Verify repository exists
    if _, err := os.Stat(filepath.Join(repoPath, "HEAD")); err != nil {
        t.Errorf("Repository not properly initialized: %v", err)
    }
    
    // Test invalid repository name
    _, err = CreateRepository("invalid/repo/name")
    if err == nil {
        t.Error("Expected error for invalid repository name")
    }
}
```

**Integration Test Example**
```go
func TestGitPushIntegration(t *testing.T) {
    // Setup test server
    server := httptest.NewServer(setupTestRouter())
    defer server.Close()
    
    // Parse server URL
    url, _ := url.Parse(server.URL)
    repoURL := fmt.Sprintf("%s/test-repo.git", url)
    
    // Initialize a test repository
    tempDir := t.TempDir()
    localRepo := filepath.Join(tempDir, "test-repo")
    
    // Run git init
    cmd := exec.Command("git", "init", localRepo)
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to initialize test repo: %v", err)
    }
    
    // Add a test file
    testFile := filepath.Join(localRepo, "test.txt")
    if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    // Add and commit
    cmd = exec.Command("git", "-C", localRepo, "add", ".")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to add files: %v", err)
    }
    
    cmd = exec.Command("git", "-C", localRepo, "commit", "-m", "Initial commit")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to commit: %v", err)
    }
    
    // Push to test server
    cmd = exec.Command("git", "-C", localRepo, "push", repoURL, "master")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to push: %v", err)
    }
}
```

## Deployment

### Building for Release

```bash
# Build for all architectures
for arch in amd64 armv7 aarch64 i386; do
    GOARCH=$arch go build -o mygit-$arch ./src/main.go
    tar czf mygit-$arch.tar.gz mygit-$arch
    rm mygit-$arch
done
```

### Creating Home Assistant Addon

```bash
# Build the addon image
podman build -t ghcr.io/flozsc/mygit -f Dockerfile .

# Push to container registry
podman push ghcr.io/flozsc/mygit

# Update addon repository
cd ..
git add mygit
git commit -m "Update mygit addon"
git push origin main
```

### Versioning

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Debugging

### Common Issues

**Port already in use**
```bash
# Find process using port
sudo lsof -i :3000

# Kill process
kill -9 <PID>
```

**Permission denied**
```bash
# Check directory permissions
ls -la /data/repos

# Fix permissions
chown -R git:git /data/repos
chmod -R 755 /data/repos
```

**Database corruption**
```bash
# Backup database
cp /data/mygit.db /data/mygit.db.backup

# Rebuild database
rm /data/mygit.db
# Restart addon (will recreate database)
```

### Logging

```bash
# View addon logs
podman logs mygit

# Follow logs
podman logs -f mygit

# Increase log level
export LOG_LEVEL=debug
./mygit
```

## Contributing

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Review Guidelines

- All changes must include tests
- Maintain backward compatibility where possible
- Follow existing code style
- Update documentation as needed
- Keep changes focused and atomic

### Issue Tracking

- Use GitHub Issues for bug reports and feature requests
- Include reproduction steps for bugs
- Provide context for feature requests
- Use labels appropriately

## Documentation

### Writing Docs

- Use Markdown format
- Keep documentation up-to-date
- Include code examples
- Explain the "why" not just the "how"

### Doc Structure

```
docs/
├── USER_GUIDE.md        # User-facing documentation
├── ADMIN_GUIDE.md       # Administration guide
├── API_REFERENCE.md     # API documentation
├── ARCHITECTURE.md      # Architecture overview
├── DEVELOPMENT.md       # This file
├── SECURITY.md          # Security practices
└── ROADMAP.md           # Future plans
```

## Best Practices

### Performance

- Minimize allocations in hot paths
- Use connection pooling
- Cache frequently accessed data
- Optimize database queries
- Use efficient data structures

### Security

- Validate all inputs
- Use prepared statements
- Encrypt sensitive data
- Implement rate limiting
- Keep dependencies updated
- Follow principle of least privilege

### Maintainability

- Write clear, focused functions
- Use descriptive names
- Add comprehensive comments
- Keep functions short (< 50 lines)
- Avoid global state
- Write tests for all code paths

## Future Development

### Roadmap

```
Q2 2026
├── OAuth integration
├── Webhook system
└── Basic CI/CD

Q3 2026
├── User management UI
├── Repository templates
└── Performance optimizations

Q4 2026
├── Plugin system
├── Federation support
└── Mobile app
```

### Experimental Features

- **AI Assist**: AI-powered code suggestions
- **Collaboration**: Real-time editing
- **Packages**: Package registry
- **Wiki**: Built-in documentation

## Conclusion

This development guide provides a comprehensive overview of building, testing, and maintaining MyGit. The project is designed to be approachable for new contributors while maintaining high standards for code quality and security.