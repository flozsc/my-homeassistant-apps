# Gitea

Gitea is a painless self-hosted Git service. This Home Assistant app runs Gitea with SQLite database storage and configurable admin credentials.

## Configuration

- **Root URL**: The external URL for accessing Gitea (e.g., `http://192.168.1.100:3000`)
- **Admin Username**: Username for the administrator account
- **Admin Email**: Email address for the administrator account
- **Admin Password**: Password for the administrator account
- **HTTP Port**: Port for the web interface (default: 3000)
- **SSH Port**: Port for SSH git operations (default: 2222)

## Access

After installation:
1. Navigate to `http://<your-ha-ip>:3000` in your browser
2. Log in with the configured admin credentials
3. Start creating repositories

Registration is disabled by default. Only the admin account can create new users.

## Storage

- **Database**: SQLite database stored at `/data/gitea/gitea.db`
- **Repositories**: Git data stored at `/data/gitea`

Both are persisted in the Home Assistant app data directory.
