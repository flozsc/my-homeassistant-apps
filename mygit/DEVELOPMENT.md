# MyGit Development Guide

## 🎯 Development Approach

**Key Decision**: Simple and reliable - use Home Assistant's built-in S6 overlay without custom configuration.

### Why This Approach
- Avoids PID 1 conflicts
- Uses proven HA patterns
- Simpler to maintain
- Works reliably in production

### When to Change
Only reconsider if:
- Specific S6 features are absolutely required
- Home Assistant changes base image significantly
- Performance issues arise that require custom S6 tuning

## 🚀 Getting Started

### Prerequisites

- **Go 1.25+** (required for local development)
- **Git** (required for version control)
- **Docker/Podman** (optional, for container testing)

### Install Go

```bash
# Download and install Go 1.26 (recommended)
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz

# Add to your PATH
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

## 🔨 Local Development

### Quick Start

```bash
# Clone the repository
git clone https://github.com/flozsc/my-homeassistant-apps.git
cd my-homeassistant-apps/mygit

# Build the application
go build -o mygit ./src/main.go

# Run with environment variables
HTTP_PORT=3000 \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD=your_secure_password \
REPO_STORAGE=./local-repos \
./mygit
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_PORT` | No | 3000 | Port to listen on |
| `ADMIN_USERNAME` | No | admin | Admin username |
| `ADMIN_PASSWORD` | No | (none) | Admin password (REQUIRED for production) |
| `REPO_STORAGE` | No | ./local-repos | Repository storage path |

**Important**: Always set `ADMIN_PASSWORD` in production. Local testing uses default `admin:admin` credentials if not set.

### Using the Test Script

```bash
# Run comprehensive local tests
./test-local.sh

# Run with custom password
ADMIN_PASSWORD=yourpass ./test-local.sh

# Run on custom port
HTTP_PORT=8080 ./test-local.sh
```

## 🧪 Testing

### Key Testing Principle
**Test in the target environment first** - Home Assistant behavior differs from local testing.

### Home Assistant Environment

The add-on uses **bashio framework** for configuration:

```yaml
# config.yaml options
http_port: 3000
admin_username: admin
admin_password: null  # Set via HA UI
repo_storage: /data/repos
```

### Local Testing

**Purpose**: Quick development and debugging
**Limitation**: Behavior may differ from Home Assistant environment

```bash
# Start server
HTTP_PORT=3000 ADMIN_USERNAME=admin REPO_STORAGE=./test-repos ./mygit &

# Test endpoints
curl -u admin:admin http://localhost:3000/
curl -u admin:admin http://localhost:3000/repos

# Test Git operations
git clone http://localhost:3000/test.git
cd test
echo "test" > README.md
git add . && git commit -m "test"
git push origin main
```

**Important**: Always verify fixes in Home Assistant environment, not just locally.

## 🐳 Container Development

### Build Container

```bash
podman build -t mygit-dev -f Dockerfile .
```

### Run Container

```bash
podman run --rm -it \
  -p 3000:3000 \
  -e HTTP_PORT=3000 \
  -e ADMIN_USERNAME=admin \
  -e REPO_STORAGE=/data/repos \
  localhost/mygit-dev
```

## 🔧 Common Issues & Solutions

### S6 Overlay Issues

**Symptom**: `s6-overlay-suexec: fatal: unable to setgid to root: Operation not permitted`

**Solution**: Ensure `init: true` in `config.yaml` and proper S6 service configuration.

### Permission Problems

**Symptom**: `Permission denied` when accessing repositories

**Solutions**:
- Ensure `REPO_STORAGE` directory exists and is writable
- Run as non-root user (git:git in container)
- Check SELinux context: `chcon -R -t container_file_t /path/to/repos`

### Authentication Failures

**Symptom**: `401 Unauthorized` even with correct credentials

**Solutions**:
- Verify `ADMIN_PASSWORD` is set in Home Assistant config
- Check password is not empty/null
- Test with default `admin:admin` credentials first

### Port Conflicts

**Symptom**: `Address already in use`

**Solutions**:
- Change `HTTP_PORT` to unused port
- Kill existing process: `lsof -i :3000` then `kill <PID>`
- Use different port for testing

### Go Module Issues

**Symptom**: `cannot find module` or dependency errors

**Solutions**:
- Run `go mod tidy`
- Delete `go.mod` and `go.sum`, then regenerate
- Ensure Go version matches project requirements (1.25+)

### Root vs Non-Root Issues

**Symptom**: Container fails with permission errors

**Solutions**:
- Run as dedicated user (git:git)
- Avoid running as root in production
- Set proper file permissions: `chown -R git:git /data/repos`

## 📚 Architecture

### Component Overview

```
Client → Nginx Proxy → MyGit → Git Backend
                       ↓
                  Auth Middleware
                       ↓
                  Config Manager
```

### Key Components

- **HTTP Server**: Handles web and Git Smart HTTP requests
- **Auth Middleware**: Basic auth and API key authentication
- **Git Backend**: Manages repository operations
- **Config Manager**: Handles add-on configuration

## 🤝 Contributing

### Code Style

- Follow Go standards (`gofmt`)
- Use descriptive commit messages
- Add tests for new features
- Document public APIs

### Pull Requests

1. Fork the repository
2. Create feature branch
3. Commit changes
4. Push and create PR
5. Wait for review

### Development Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes, commit
git add .
git commit -m "Add feature: brief description"

# Push to your fork
git push origin feature/your-feature

# Create PR to main branch
```

## 🔄 Avoiding the Testing Loop

### Common Pitfall
Getting stuck in a cycle of:
1. Test locally → Add configuration to make it work
2. Deploy to HA → Breaks because HA environment differs
3. Test locally again → Add more configuration
4. Repeat...

### Solution
- **Test in target environment first** (Home Assistant)
- **Keep configuration minimal**
- **Remove old configuration completely** when changing approaches
- **Document decisions clearly** to avoid confusion

### When to Test Locally
- Quick development and debugging
- Testing basic functionality
- Performance optimization

### When to Test in Home Assistant
- Before releasing any version
- When changing process management
- When modifying configuration handling
- Always for final verification

## 📊 Debugging

### Log Locations

- **Home Assistant**: Add-on logs in HA UI
- **Local**: Console output
- **Container**: `podman logs <container>`

### Debug Commands

```bash
# Check running processes
ps aux | grep mygit

# Check listening ports
netstat -tlnp | grep 3000

# Check Git configuration
git config --list

# Test Git protocol manually
curl -v http://localhost:3000/test.git/info/refs?service=git-upload-pack
```

## 🔮 Advanced Topics

### Custom Git Hooks

Add hooks to `repo_storage/.git/hooks/` directory.

### Performance Tuning

- Adjust Go runtime settings
- Optimize Git operations
- Configure proper timeouts

### Monitoring

- Add health check endpoint
- Implement metrics collection
- Set up logging to file

## 📞 Support

For issues:
- Check GitHub issues
- Review documentation
- Ask in Home Assistant forums
- Open new issue with details

Include:
- MyGit version
- Home Assistant version
- Error logs
- Configuration
- Steps to reproduce