# MyGit - Private Git Server for Home Assistant

A lightweight, private Git server designed to run as a Home Assistant addon.

## Features

### ✅ Core Features
- **Private Git Repository Hosting**: Host your own Git repositories
- **Web Interface**: Browse repositories through a simple web UI
- **Authentication**: Basic Auth and API key support
- **SSH Access**: Push/pull over SSH with key authentication
- **Home Assistant Integration**: Native addon with Supervisor API support

### 🚧 Roadmap (Future Features)
- **OAuth Integration**: GitHub/GitLab/OAuth providers
- **Webhooks**: Trigger actions on push events
- **Simple CI/CD**: Basic build automation
- **User Management**: Multiple users with different permissions
- **Repository Statistics**: Activity charts and insights
- **Mobile Optimization**: Touch-friendly interface

## Installation

### As Home Assistant Addon
1. Add this repository to your Home Assistant
2. Install the "MyGit" addon
3. Configure options in the addon settings
4. Start the addon

### Manual Installation
```bash
# Clone the repository
git clone https://github.com/flozsc/mygit
cd mygit

# Build the application
podman build -t mygit -f Dockerfile .

# Run the container
podman run -d \
  -p 3000:3000 \
  -p 2222:22 \
  -v $(pwd)/data:/data \
  mygit
```

## Configuration

### Addon Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `http_port` | number | 3000 | HTTP port for web interface |
| `admin_username` | string | "admin" | Admin username |
| `admin_password` | password | - | Admin password (set on first run) |
| `repo_storage` | string | "/data/repos" | Repository storage path |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `HTTP_PORT` | HTTP port (overrides config) |
| `ADMIN_USERNAME` | Admin username (overrides config) |
| `ADMIN_PASSWORD` | Admin password (overrides config) |
| `REPO_STORAGE` | Repository storage path (overrides config) |

## Authentication

### API Keys
Generate API keys in the web interface for programmatic access:

```bash
# Use API key for authentication
curl -H "Authorization: Bearer YOUR_API_KEY" http://mygit:3000/api/repos
```

### SSH Keys
Add SSH public keys through the web interface for Git operations:

```bash
# Clone over SSH
git clone ssh://git@mygit:2222/your-repo.git
```

### Basic Auth
Use your admin credentials for web interface access.

## Usage

### Creating a Repository
```bash
# Via Git push (auto-creates repository)
git push ssh://git@mygit:2222/new-repo.git master

# Via Web UI
# 1. Open http://mygit:3000
# 2. Click "New Repository"
# 3. Enter repository name
# 4. Click "Create"
```

### Cloning a Repository
```bash
# Over HTTP (read-only)
git clone http://mygit:3000/your-repo.git

# Over SSH (read-write)
git clone ssh://git@mygit:2222/your-repo.git
```

## Development

### Prerequisites
- Go 1.20+
- Docker or Podman
- Make (optional)

### Building
```bash
# Build the application
make build

# Run tests
make test

# Build Docker image
make docker
```

### Project Structure
```
mygit/
├── src/                # Go source code
│   ├── git/             # Git operations
│   ├── web/             # Web interface
│   ├── auth/            # Authentication
│   └── config/          # Configuration
├── web/                # Web assets
│   ├── static/         # CSS, JS, images
│   └── templates/       # HTML templates
├── tests/               # Test suite
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── e2e/             # End-to-end tests
├── docs/                # Documentation
└── Dockerfile           # Container definition
```

## Security

### Best Practices
- All credentials are stored encrypted
- Repository data is isolated
- Regular security updates
- Rate limiting on all endpoints
- CSRF protection for web UI

### AppArmor Profile
The addon includes a comprehensive AppArmor profile for enhanced security.

## Troubleshooting

### Common Issues

**"Permission denied" when pushing**
- Ensure your SSH key is added to your account
- Check repository permissions
- Verify the repository exists

**Web interface not loading**
- Check the addon logs
- Verify the HTTP port is not conflicting
- Ensure the addon is running

**Out of disk space**
- Clean up old repositories
- Check your storage configuration
- Monitor disk usage

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to your branch
5. Open a pull request

### Code Standards
- Go formatting (`gofmt`)
- Comprehensive tests
- Clear documentation
- Semantic commits

## License

MIT License. See `LICENSE` for details.

## Support

For issues, questions, or feature requests:
- Open an issue on GitHub
- Check the discussion forums
- Review the documentation

---

**MyGit** - Your private Git server, built for Home Assistant! 🚀