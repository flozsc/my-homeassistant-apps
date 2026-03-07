# MyGit Development Guidelines

## 🎯 Project Vision
**MyGit** is a lightweight, private Git server designed specifically for Home Assistant. It provides:
- Simple repository hosting
- Basic web UI for browsing code
- User management (private access)
- Seamless Home Assistant integration

## 🤖 Agent Guidelines

### 1. Development Principles
- **Keep It Simple**: Focus on core functionality first
- **Home Assistant First**: Design for the addon ecosystem
- **Security by Default**: Private by default, explicit sharing
- **Documentation Driven**: Document before implementing
- **Test Early, Test Often**: Build tests alongside features

### 2. Technical Standards

#### Go Code
```go
// Example of preferred Go style
func CreateRepository(name string) (string, error) {
    // Validate input
    if !isValidRepoName(name) {
        return "", fmt.Errorf("invalid repository name: %s", name)
    }

    // Create repository
    repoPath := filepath.Join(repoStorage, name+".git")
    if err := os.MkdirAll(repoPath, 0755); err != nil {
        return "", fmt.Errorf("failed to create repo: %w", err)
    }

    return repoPath, nil
}
```

#### File Structure
```
mygit/
├── src/                # Go source code
├── web/                # Web interface
├── tests/              # Test suite
├── docs/               # Documentation
└── Dockerfile          # Container definition
```

### 3. Home Assistant Integration
- Use `init: true` for S6 overlay compatibility
- Keep configuration simple - use environment variables
- Follow addon configuration best practices
- Implement health checks
- Support backup/restore
- Use Supervisor API for configuration
- Let Home Assistant's built-in S6 handle process management

### 4. Authentication Flow
```
Client → Authentication Middleware → Permission Check → Handler
```
- Support API keys, SSH keys, and Basic Auth
- Store credentials securely
- Implement rate limiting
- Use HTTPS for all communications

### 5. Git Protocol Implementation
- Start with **smart HTTP** protocol
- Add **SSH protocol** support
- Implement **receive-pack** for pushes
- Support **upload-pack** for fetches
- Handle **repository auto-creation**

### 6. Testing Strategy
- **Unit Tests**: 80%+ coverage for core modules
- **Integration Tests**: Component interactions
- **E2E Tests**: Full workflow validation
- **Security Tests**: Penetration testing
- **Performance Tests**: Benchmarking

### 7. Documentation Standards
- **USER_GUIDE.md**: User-facing documentation
- **ARCHITECTURE.md**: Technical overview
- **DEVELOPMENT.md**: Development guide
- **API_REFERENCE.md**: API documentation
- **Code Comments**: Comprehensive inline docs

### 8. Lessons Learned from Gitea Issues
- **S6 Overlay**: Use Home Assistant's built-in S6, don't add custom configuration
- **Simplicity**: Keep process management simple and reliable
- **Permissions**: Volume mounts need careful handling
- **Non-Root**: Run as dedicated user (git:git)
- **Configuration**: Use environment variables with Supervisor API fallbacks
- **Testing**: Test in target environment (Home Assistant), not just locally
- **PID 1**: Let Home Assistant's S6 be PID 1, our app runs as child process

### 9. Decision Making Principles
1. **Favor Simplicity**: When in doubt, choose the simpler solution
2. **Home Assistant Compatibility**: Always prioritize integration
3. **Security First**: Never compromise on security
4. **Documentation**: If it's not documented, it doesn't exist
5. **Test Coverage**: No feature is complete without tests

### 9. S6 Overlay Decision

**Approach**: Simple and reliable - use Home Assistant's built-in S6 overlay without custom configuration

**Rationale**:
- Avoids PID 1 conflicts and complexity
- Uses Home Assistant's proven S6 implementation
- Simpler to maintain and debug
- Works reliably in production

**Implementation**:
- `init: true` in config.yaml
- Simple run script as entrypoint
- No custom S6 service files
- Environment variables for configuration

**When to Revisit**:
- If specific S6 features are needed
- If Home Assistant changes base image significantly
- If performance issues arise

### 9. S6 Overlay Decision

**Approach**: Use Home Assistant base image's S6 overlay with `init: false`

**Rationale**:
- Home Assistant base image (v3+) includes S6 overlay
- Setting `init: false` prevents double initialization
- Follows official HA documentation (configuration.md line 195)
- Avoids PID 1 conflicts completely
- Simpler and more reliable than custom S6 configuration

**Implementation**:
- `init: false` in config.yaml (CRITICAL)
- Simple run script as entrypoint
- Minimal S6 service wrapper
- Environment variables for configuration

**Official Documentation**:
> "Starting in V3 of S6 setting this to `false` is required or the addon won't start"
> - Home Assistant Apps Configuration Documentation

**When to Revisit**:
- If Home Assistant changes base image significantly
- If official recommendations change
- If specific init features become necessary

### 10. Future Roadmap
```
Q2 2026
├── OAuth integration
├── Webhook system
└── Basic CI/CD

Q3 2026
├── User management UI
├── Repository templates
└── Performance optimizations
```

## 📋 Implementation Checklist

### For New Features
- [ ] Update architecture documentation
- [ ] Write unit tests first
- [ ] Implement feature
- [ ] Add integration tests
- [ ] Update user documentation
- [ ] Test locally with podman
- [ ] Verify Home Assistant integration

### For Bug Fixes
- [ ] Reproduce the issue
- [ ] Write regression test
- [ ] Implement fix
- [ ] Verify in multiple environments
- [ ] Update documentation if needed

### For All Code Changes
- [ ] Commit changes to GitHub with a descriptive commit message
- [ ] Push changes to the appropriate branch (usually main)
- [ ] Verify changes are visible in the GitHub repository
- [ ] Update any related documentation or changelogs

## 🤝 Collaboration Guidelines

### When Working with Users
1. **Ask Questions**: Clarify requirements before implementing
2. **Provide Options**: Present tradeoffs when appropriate
3. **Document Decisions**: Record why choices were made
4. **Validate Understanding**: Summarize plans before executing
5. **Iterate**: Start small, get feedback, improve

### When Making Changes
1. **Read First**: Understand existing code
2. **Plan**: Create a well-researched plan
3. **Test**: Verify changes work
4. **Document**: Update all relevant docs
5. **Review**: Present changes for feedback

### Commit Policy
1. **Always Commit**: All code changes must be committed to GitHub before considering work complete
2. **Descriptive Messages**: Use clear, descriptive commit messages that explain the "why" not just the "what"
3. **Atomic Commits**: Make small, focused commits for logical units of change
4. **Version Bumps**: Always bump versions when making breaking changes or significant updates
5. **Tag Releases**: Create Git tags for major releases and versions