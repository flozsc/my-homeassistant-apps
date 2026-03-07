# MyGit Development Guidelines

## Project Overview
**MyGit** is a lightweight, private Git server for Home Assistant (HA Apps).
- Simple repository hosting over Git HTTP
- Web UI for browsing repositories
- Basic Auth authentication
- HA addon/App integration

---

## Development Commands

### Local Development (Preferred for Testing)
```bash
cd mygit

# Build and run (uses web templates from ./web/)
go build -o mygit ./src/main.go

# Run with environment variables
HTTP_PORT=3000 \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD=secret \
REPO_STORAGE=./test-repos \
./mygit

# Or use the test script (builds + runs + tests)
ADMIN_PASSWORD=secret ./test-local.sh
```

### Single Test
```bash
go test -v ./src/... -run TestName
```

### All Tests
```bash
go test -v ./src/...
```

### Lint & Format
```bash
gofmt -w ./src/
go vet ./src/...
```

### Build for HA
The Dockerfile handles building inside the container - just commit and HA rebuilds.

---

## Code Style Guidelines

### Go Conventions
- Use `gofmt` for formatting
- camelCase for functions/variables, PascalCase for types
- Return errors with context: `fmt.Errorf("failed to create repo: %w", err)`
- Never ignore errors with `_`

### Imports
Standard library first, then external packages (alphabetical):
```go
import (
    "fmt"
    "net/http"
    "os"

    "github.com/flozsc/mygit/src/auth"
)
```

### Types
- Use explicit types
- Prefer interfaces for testability
- Document all exported functions

### Error Handling
- Return errors, don't log and continue
- Use wrapped errors for context: `%w`
- Handle errors explicitly

---

## Home Assistant Integration

### S6 Overlay (Critical)
- Use `init: false` in config.yaml
- Add minimal S6 service at `/etc/services.d/mygit/run`:
  ```bash
  #!/bin/sh
  exec /run.sh
  ```
- HA's built-in S6 handles process management

### Config Flow
1. config.yaml defines options (http_port, admin_username, etc.)
2. run.sh reads via environment variables
3. Go app reads from environment

### Required Files
- `config.yaml` - HA addon config
- `Dockerfile` - Container build
- `run.sh` - Entry point
- `rootfs/etc/services.d/mygit/run` - S6 service definition
- `web/templates/` - HTML templates (copied to /data/web/)
- `web/static/` - CSS/images (copied to /data/web/static/)

---

## Testing Strategy
1. **Local first**: Test changes locally with `./test-local.sh`
2. **HA second**: Rebuild addon in HA for final verification
3. **Git operations**: Test clone/push to verify Git HTTP works

---

## Project Structure
```
mygit/
├── src/
│   ├── main.go           # Main app + handlers
│   └── auth/             # Authentication middleware
├── web/
│   ├── templates/        # HTML templates
│   └── static/           # CSS, favicon
├── config.yaml           # HA addon config
├── Dockerfile            # Container build
├── run.sh               # Entry point
├── rootfs/              # S6 service files
└── test-local.sh        # Local development script
```

---

## Common Issues

### Port Not Accessible
- Ensure `ports: 3000/tcp: 3000` in config.yaml

### App Not Starting
- Check S6 service file exists at `/etc/services.d/mygit/run`
- Check run.sh is executable

### Web UI Not Rendering
- Verify web/ files copied to /data/web/ in Dockerfile
- Check templates parse correctly

---

## Commit Policy
1. Always commit changes to GitHub
2. Use descriptive commit messages
3. Bump version in config.yaml for significant changes
4. Test locally before pushing
