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
- Implement proper S6 service configuration with `/etc/services.d/` directory
- Follow addon configuration best practices
- Implement health checks
- Support backup/restore
- Use Supervisor API for configuration
- Ensure applications run as child processes of S6 (not PID 1)

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
- **S6 Overlay**: Must run as PID 1, configure properly with service directories
- **S6 Service Files**: Require `/etc/services.d/<service>/run` and `finish` scripts
- **Permissions**: Volume mounts need careful handling
- **Non-Root**: Run as dedicated user (git:git)
- **Configuration**: Use Supervisor API with fallbacks
- **Testing**: Local testing with podman works well
- **PID 1**: S6 overlay must be PID 1, applications run as child processes

### 9. Decision Making Principles
1. **Favor Simplicity**: When in doubt, choose the simpler solution
2. **Home Assistant Compatibility**: Always prioritize integration
3. **Security First**: Never compromise on security
4. **Documentation**: If it's not documented, it doesn't exist
5. **Test Coverage**: No feature is complete without tests

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