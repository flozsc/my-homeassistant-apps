# Home Assistant Add-on: Gitea

A Home Assistant add-on to run [Gitea](https://gitea.io/) - a self-hosted Git service written in Go.

## About

This add-on runs Gitea in a Docker container managed by Home Assistant. It's perfect for hosting your own Git repositories privately or for hobby projects.

## Features

- Self-hosted Git service
- Full web UI with repository management
- User authentication
- Optional SSH access for Git operations
- Configurable external URL for reverse proxy setup (e.g., Nginx Proxy Manager)
- Data persists in Home Assistant's `/share` directory

## Installation

1. Add this repository to Home Assistant:
   - Go to **Settings** → **Add-ons** → **Add-on store**
   - Click the **⋮** menu → **Add repository**
   - Enter: `https://github.com/flozsc/my-homeassistant-apps`

2. Install the **Gitea** add-on

3. Configure your settings (see Configuration below)

4. Start the add-on

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `hostname` | string | `gitea.local` | Server hostname |
| `http_port` | int | `3000` | HTTP port |
| `ssh_port` | int | `2222` | SSH port (set to `0` to disable) |
| `root_url` | string | - | External URL (e.g., `https://gitea.yourdomain.com`) for reverse proxy setups |
| `admin_password` | password | - | Set admin password on first run (optional - leave blank to keep existing) |

### First Setup

1. After starting the add-on for the first time, open the web UI
2. Default admin username: **gitea_admin**
3. Set your admin password in the add-on configuration to create the admin user

### Reverse Proxy Setup

If using Nginx Proxy Manager or another reverse proxy:

1. Set `root_url` to your external URL (e.g., `https://gitea.yourdomain.com`)
2. Configure your reverse proxy to forward to `http://<home-assistant-ip>:3000`

## Data Location

- Repository data: `/share/gitea`
- Database: `/share/gitea/gitea.db`

## Disclaimer

This is my personal hobby project built with "vibe coding" - I make things work without fully understanding them, relying on intuition and experimentation. It works for my use case, but may not be production-ready or follow best practices.

**Use at your own risk!**

That said, if you find issues or have improvements, feel free to submit pull requests. Contributions are welcome!

## License

MIT
