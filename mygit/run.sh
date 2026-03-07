#!/usr/bin/with-contenv bashio

# Read configuration
HTTP_PORT=$(bashio::config 'http_port')
ADMIN_USERNAME=$(bashio::config 'admin_username')
REPO_STORAGE=$(bashio::config 'repo_storage')

# Export for the application
export HTTP_PORT
export REPO_STORAGE
export ADMIN_USERNAME

# If admin password is set, export it
if bashio::config.exists 'admin_password'; then
    ADMIN_PASSWORD=$(bashio::config 'admin_password')
    export ADMIN_PASSWORD
fi

# Ensure repository directory exists
mkdir -p "$REPO_STORAGE"
chmod 755 "$REPO_STORAGE"

# Start the application
bashio::log.info "Starting MyGit v1.0.0..."
bashio::log.info "  HTTP Port: $HTTP_PORT"
bashio::log.info "  Repository Storage: $REPO_STORAGE"
bashio::log.info "  Admin User: $ADMIN_USERNAME"

if [ -n "$ADMIN_PASSWORD" ]; then
    bashio::log.info "  Admin Password: [SET]"
else
    bashio::log.warning "  Admin Password: [NOT SET - Web UI will be read-only]"
fi

exec /usr/local/bin/mygit