# MyGit Architecture

## Overview

MyGit is designed as a lightweight, self-contained Git server that integrates seamlessly with Home Assistant. The architecture follows modern Go best practices and emphasizes security, simplicity, and performance.

## System Architecture

```
┌───────────────────────────────────────────────────────┐
│                    MyGit Application                   │
├───────────────────┬───────────────────┬───────────────┤
│   Web Interface   │    Git Server     │  Auth System  │
└─────────┬─────────┴─────────┬─────────┴───────┬───────┘
          │                   │                   │
          ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────┐
│  HTTP Server    │ │  Git Protocol   │ │  API Keys   │
│  (Port 3000)    │ │  (Port 2222)   │ │  SSH Keys   │
└─────────────────┘ └─────────────────┘ └─────────────┘
          │                   │                   │
          ▼                   ▼                   ▼
┌───────────────────────────────────────────────────────┐
│                   Data Storage                        │
│  ┌─────────────┐    ┌─────────────────────────────┐  │
│  │ SQLite DB  │    │  Repository Storage        │  │
│  │ (Metadata) │    │  (Git repositories)        │  │
│  └─────────────┘    └─────────────────────────────┘  │
└───────────────────────────────────────────────────────┘
```

## Component Details

### 1. Web Interface

**Technology**: Go HTTP server with HTML templates

**Components**:
- **Router**: Handles URL routing and middleware
- **Templates**: HTML templates with Go's `html/template`
- **API**: RESTful API endpoints
- **Assets**: Static files (CSS, JS, images)

**Features**:
- Repository browser
- Commit history viewer
- User management
- Settings interface

### 2. Git Server

**Technology**: Direct Git command execution

**Components**:
- **HTTP Transport**: Git over HTTP (smart protocol)
- **SSH Transport**: Git over SSH
- **Repository Manager**: Create/delete/manage repos
- **Hook System**: Pre/post-receive hooks

**Features**:
- Smart HTTP protocol support
- SSH key authentication
- Repository auto-creation on push
- Access control

### 3. Authentication System

**Technology**: JWT and SSH key management

**Components**:
- **API Key Manager**: Generate/revoke API keys
- **SSH Key Manager**: Manage authorized keys
- **Session Manager**: Web session handling
- **Permission System**: Role-based access control

**Features**:
- API key authentication
- SSH public key authentication
- Basic Auth for web interface
- Rate limiting

### 4. Storage Layer

**Technology**: SQLite + Filesystem

**Components**:
- **SQLite Database**: User metadata, API keys, settings
- **Repository Storage**: Git repositories on disk
- **Cache Layer**: In-memory caching for performance

**Features**:
- Efficient metadata storage
- Repository isolation
- Backup-friendly structure

## Data Flow

### HTTP Request Flow

```
Client → Nginx (reverse proxy) → MyGit HTTP Server → Authentication → Authorization → Handler → Response
```

### Git Push Flow

```
Git Client → SSH Server → Authentication → Git Receive-Pack → Repository Update → Hooks → Response
```

### Web UI Flow

```
Browser → HTTP Server → Session Check → Template Rendering → Database Query → Response
```

## Security Architecture

### Authentication Layers

1. **Transport Layer**: TLS encryption (via Home Assistant ingress)
2. **Authentication**: API keys, SSH keys, or Basic Auth
3. **Authorization**: Role-based access control
4. **Rate Limiting**: Protection against brute force

### Security Measures

- **AppArmor Profile**: Restricts system calls
- **Non-Root Operation**: Runs as dedicated `git` user
- **Input Validation**: All user input is validated
- **SQLite Encryption**: Sensitive data encrypted at rest
- **CSRF Protection**: Web forms protected
- **Security Headers**: CSP, HSTS, etc.

## Performance Considerations

### Optimization Strategies

1. **Caching**: In-memory cache for frequently accessed data
2. **Connection Pooling**: Reuse database connections
3. **Efficient Git Operations**: Minimize disk I/O
4. **Static Asset Caching**: Long cache headers for static files
5. **Gzip Compression**: Reduce bandwidth usage

### Scalability

- **Horizontal Scaling**: Not designed for clustering (single-node focus)
- **Vertical Scaling**: Optimized for resource-constrained environments
- **Repository Limits**: Designed for personal/hobby use (100s of repos)

## Integration with Home Assistant

### Addon Lifecycle

```
Install → Configure → Start → Monitor → Update → Backup/Restore
```

### Supervisor Integration

- **API**: Uses Supervisor API for configuration
- **Health Checks**: Regular status reporting
- **Logging**: Integrated with Supervisor logs
- **Backups**: Included in Home Assistant backups

### Network Integration

- **Ingress**: Optional web UI through Home Assistant frontend
- **Ports**: HTTP (3000), SSH (2222)
- **DNS**: Automatic hostname resolution

## Future Architecture Evolution

### Planned Enhancements

1. **OAuth Integration**: External identity providers
2. **Webhook System**: Event-driven notifications
3. **CI/CD Pipelines**: Basic build automation
4. **Federation**: Cross-instance repository access
5. **Plugins**: Extensible architecture

### Scalability Improvements

1. **Repository Sharding**: Distribute large repos
2. **Read Replicas**: For high-traffic instances
3. **Object Storage**: For large files
4. **CDN Integration**: For static assets

## Development Guidelines

### Coding Standards

- **Go**: Follow `gofmt` and `golint` standards
- **HTML**: Semantic markup, accessible
- **CSS**: BEM methodology
- **JavaScript**: ES6+, minimal dependencies

### Testing Strategy

- **Unit Tests**: 80%+ coverage for core modules
- **Integration Tests**: Component interactions
- **E2E Tests**: Full workflow validation
- **Security Tests**: Penetration testing
- **Performance Tests**: Benchmarking

### Documentation

- **Code Comments**: Comprehensive inline documentation
- **API Docs**: Swagger/OpenAPI specification
- **User Guide**: Complete usage documentation
- **Architecture Docs**: This document

## Deployment Topologies

### Single Node (Recommended)

```
┌───────────────────────────────────────┐
│           Home Assistant Server        │
│                                       │
│  ┌─────────────┐    ┌───────────────┐  │
│  │  MyGit      │    │ Other Addons  │  │
│  │  (Container)│    │  (Containers)│  │
│  └─────────────┘    └───────────────┘  │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │         Supervisor              │  │
│  └─────────────────────────────────┘  │
└───────────────────────────────────────┘
```

### Multi-Node (Advanced)

```
┌─────────────┐    ┌─────────────┐
│ HA Node 1   │    │ HA Node 2   │
│ (Primary)   │    │ (Backup)   │
└──────┬──────┘    └──────┬──────┘
       │                   │
       ▼                   ▼
┌───────────────────────────────────┐
│           Shared Storage          │
│  (NFS, Ceph, or similar)         │
└───────────────────────────────────┘
```

## Monitoring and Observability

### Metrics

- Repository count
- User count
- Request rates
- Error rates
- Response times
- Disk usage

### Logging

- Structured logs (JSON)
- Log levels (DEBUG, INFO, WARN, ERROR)
- Log rotation
- Supervisor integration

### Alerting

- Disk space warnings
- Authentication failures
- Rate limit triggers
- Server errors

## Backup and Recovery

### Backup Strategy

- **Full Backups**: Complete repository and database
- **Incremental**: Git bundles for large repos
- **Export/Import**: Standard Git bundle format

### Recovery Procedures

1. Restore from Home Assistant backup
2. Manual repository recovery from Git bundles
3. Database recovery from SQLite dump
4. Emergency access procedures

## Conclusion

MyGit's architecture is designed for simplicity, security, and seamless integration with Home Assistant. The modular design allows for easy maintenance and future expansion while maintaining a small footprint suitable for home servers and hobby use.