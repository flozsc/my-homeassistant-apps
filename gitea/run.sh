#!/usr/bin/with-contenv bashio

bashio::log.info "Starting Gitea..."

# Try to get configuration from Supervisor API, fall back to defaults for local testing
get_config() {
    local key="$1"
    local default="$2"
    
    # Try Supervisor API first (with timeout for local testing)
    if timeout 2s bashio::config.exists "$key" >/dev/null 2>&1; then
        timeout 2s bashio::config "$key"
        return
    fi
    
    # For local testing without Supervisor, use hardcoded defaults
    echo "$default"
}

ROOT_URL=$(get_config 'root_url' 'http://localhost:3000')
ADMIN_USERNAME=$(get_config 'admin_username' 'gitea_admin')
ADMIN_EMAIL=$(get_config 'admin_email' 'admin@example.com')
HTTP_PORT=$(get_config 'http_port' '3000')
SSH_PORT=$(get_config 'ssh_port' '2222')

# Create directories if they don't exist
[ -d /data/gitea ] || mkdir -p /data/gitea
[ -d /data/git ] || mkdir -p /data/git

cat > /data/gitea/app.ini << EOF
APP_NAME = Gitea
RUN_USER = git
RUN_MODE = prod

[repository]
ROOT = /data/git/repositories

[database]
DB_TYPE = sqlite3
PATH = /data/gitea/gitea.db

[server]
DOMAIN = $(hostname -f)
HTTP_PORT = ${HTTP_PORT}
ROOT_URL = ${ROOT_URL}
SSH_PORT = ${SSH_PORT}
SSH_DOMAIN = $(hostname -f)
START_SSH_SERVER = true

[service]
DISABLE_REGISTRATION = true

[security]
INSTALL_LOCK = true
SECRET_KEY =

[admin]
DEFAULT_ADMIN_USER = ${ADMIN_USERNAME}
DEFAULT_ADMIN_EMAIL = ${ADMIN_EMAIL}
EOF

if bashio::config.exists 'admin_password'; then
    ADMIN_PASSWORD=$(bashio::config 'admin_password')
    sed -i "/DEFAULT_ADMIN_EMAIL/a DEFAULT_ADMIN_PASSWORD = ${ADMIN_PASSWORD}" /data/gitea/app.ini
    bashio::log.info "  Admin Password: [SET]"
fi

bashio::log.info "Gitea configuration:"
bashio::log.info "  Root URL: ${ROOT_URL}"
bashio::log.info "  HTTP Port: ${HTTP_PORT}"
bashio::log.info "  SSH Port: ${SSH_PORT}"
bashio::log.info "  Admin User: ${ADMIN_USERNAME}"
bashio::log.info "  Registration: Disabled"

# Ensure directories exist and have proper permissions
mkdir -p /data/gitea /data/git/repositories /data/git/custom
chmod -R 755 /data/gitea /data/git

# Run Gitea directly (let HA's S6 handle process management)
exec /usr/local/bin/gitea web --config /data/gitea/app.ini --custom-path /data/git/custom --work-path /data/git
