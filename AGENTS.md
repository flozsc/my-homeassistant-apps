# MyGit Development Guidelines

**MyGit** is a lightweight Git server for Home Assistant with HTTP Git hosting, web UI, and Basic Auth.

---

## Build & Run
```bash
cd mygit

# Build
go build -o mygit ./src/main.go

# Run with environment variables
HTTP_PORT=3000 ADMIN_USERNAME=admin ADMIN_PASSWORD=secret REPO_STORAGE=./test-repos ./mygit

# Local development script (builds + runs + tests)
ADMIN_PASSWORD=secret ./test-local.sh
```

---

## Testing
```bash
# Run all tests
go test -v ./src/...

# Run a single test
go test -v ./src/... -run TestName

# Run tests with coverage
go test -v -cover ./src/...
```

---

## Lint & Format
```bash
# Format code (always run before committing)
gofmt -w ./src/

# Run go vet
go vet ./src/...

# Check for staticcheck (if installed)
staticcheck ./src/...
```

---

## Code Style Guidelines

### Formatting
- Run `gofmt -w ./src/` before every commit
- Use 4 spaces for indentation (Go standard)
- Keep lines under 100 characters when practical

### Naming Conventions
- **Variables/Functions**: camelCase (`repoPath`, `getConfig`)
- **Types/Interfaces**: PascalCase (`Repository`, `AuthHandler`)
- **Constants**: PascalCase or camelCase (`MaxFileSize`, `httpPort`)
- **Packages**: lowercase, short (`auth`, `handlers`, `git`)
- **Files**: lowercase with underscores (`auth.go`, `git_handlers.go`)
- **Acronyms**: Keep original case (`URL`, `HTTP`, `API`)

### Imports
Standard library first, then external packages (alphabetical within groups):
```go
import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"

    "github.com/go-git/go-git/v5"
    "github.com/gorilla/mux"
)
```

### Types & Interfaces
- Use explicit types, avoid `any` unless necessary
- Prefer interfaces for testability and dependency injection
- Define interfaces close to where they're used

### Error Handling
- Always return errors with context using `%w`
- Never ignore errors with `_`
- Check errors immediately after calls
- Use sentinel errors for known conditions:
```go
var ErrRepoNotFound = errors.New("repository not found")
```

### Logging
- Use standard `log` package for simple apps
- Log at appropriate levels: `log.Printf` for info, `log.Printf("ERROR: %v", err)` for errors

### Constants & Config
- Use environment variables for all configurable values
- Provide sensible defaults
- Validate config at startup

---

## Testing Guidelines
- Tests in `*_test.go` files alongside source
- Use table-driven tests for multiple cases
- Create helper functions for common setup
- Use `t.Cleanup()` for resource cleanup

---

## Home Assistant Integration

### Required Files
- `config.yaml` - HA addon configuration
- `Dockerfile` - Container build
- `run.sh` - Entry point script
- `rootfs/etc/services.d/mygit/run` - S6 service definition

### Config Flow
1. `config.yaml` defines addon options
2. `run.sh` reads config and sets environment variables
3. Go app reads from environment

---

## Project Structure
```
mygit/
├── src/
│   ├── main.go           # Entry point + route setup
│   ├── auth/             # Authentication middleware
│   ├── handlers/         # HTTP handlers
│   ├── git/              # Git operations
│   └── config/           # Configuration loading
├── web/
│   ├── templates/        # HTML templates
│   └── static/           # CSS, images
├── config.yaml           # HA addon config
├── Dockerfile            # Container build
└── run.sh               # Entry point
```

---

## Common Issues
- **Port not accessible**: Ensure `ports: 3000/tcp: 3000` in config.yaml
- **App not starting**: Verify S6 service file at `/etc/services.d/mygit/run`, check run.sh is executable
- **Web UI not rendering**: Verify web/ files copied to /data/web/ in Dockerfile

---

## Commit Policy
1. Always run `gofmt` and `go vet` before committing
2. Test locally with `./test-local.sh` before pushing
3. Use descriptive commit messages
4. Bump version in config.yaml for significant changes
