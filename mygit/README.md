# MyGit - Private Git Server for Home Assistant

A lightweight, private Git server designed to run as a Home Assistant addon.

## Features

- **Private Git Repository Hosting**: Host your own Git repositories
- **Web Interface**: Browse repositories through a simple web UI
- **Authentication**: Basic Auth with configurable credentials
- **Git Smart HTTP**: Clone and push over HTTP
- **Home Assistant Integration**: Native addon with port exposure
- **Auto-create Repositories**: Repositories are created automatically on first push

## Installation

### As Home Assistant Addon

1. Add this repository to your Home Assistant
2. Install the "MyGit" addon
3. Configure options in the addon settings
4. Start the addon

### Manual Installation

```bash
# Clone the repository
git clone https://github.com/flozsc/my-homeassistant-apps.git
cd myhomeassistant-apps/mygit

# Build the application
podman build -t mygit -f Dockerfile .

# Run the container
podman run -d \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  mygit
```

## Configuration

### Addon Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `http_port` | number | 3000 | HTTP port for web interface |
| `admin_username` | string | "admin" | Admin username |
| `admin_password` | string | "admin" | Admin password |
| `repo_storage` | string | "/data/repos" | Repository storage path |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | 3000 | HTTP port |
| `ADMIN_USERNAME` | admin | Admin username |
| `ADMIN_PASSWORD` | admin | Admin password |
| `REPO_STORAGE` | /data/repos | Repository storage path |

## Usage

### Accessing the Web UI

Open http://<ha-ip>:3000 in your browser. You'll be prompted for credentials (default: admin/admin).

### Settings

The Settings page provides access to user management and SSH key configuration. Click "Settings" in the header navigation after logging in.

- **User Management** - Create and delete users (admin only)
- **SSH Keys** - Add or remove SSH keys for Git operations

### Creating a Repository

Repositories are created automatically when you push to them:

```bash
# Create a new repository by pushing
git init my-project
cd my-project
echo "# My Project" > README.md
git add .
git commit -m "Initial commit"
git remote add origin http://admin:admin@<ha-ip>:3000/my-project.git
git push -u origin main
```

### Cloning a Repository

```bash
# Clone over HTTP
git clone http://admin:admin@<ha-ip>:3000/your-repo.git
```

## Development

### Prerequisites

- Go 1.25+
- Docker or Podman

### Building

```bash
cd mygit
go build -o mygit ./src/main.go
```

### Testing Locally

```bash
HTTP_PORT=3000 ADMIN_USERNAME=admin ADMIN_PASSWORD=secret ./mygit
```

## Project Structure

```
mygit/
├── src/
│   ├── main.go           # Main application
│   └── auth/             # Authentication
├── web/
│   ├── templates/        # HTML templates
│   └── static/           # CSS, images
├── config.yaml           # HA addon config
├── Dockerfile            # Container definition
├── run.sh               # Entry point script
└── rootfs/              # S6 service definitions
```

## Troubleshooting

### "Permission denied" when pushing

- Ensure your credentials are correct
- Check the addon logs for more details

### Web interface not loading

- Verify the addon is running (check green indicator)
- Confirm port 3000 is not conflicting
- Check the addon logs

## License

MIT License. See `LICENSE` for details.

---

**MyGit** - Your private Git server, built for Home Assistant!
