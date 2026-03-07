#!/bin/sh

# Set default values
HTTP_PORT=${HTTP_PORT:-3000}
REPO_STORAGE=${REPO_STORAGE:-/data/repos}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}

# Export for the application
export HTTP_PORT
export REPO_STORAGE
export ADMIN_USERNAME

# If admin password is set, export it
if [ -n "${ADMIN_PASSWORD}" ]; then
    export ADMIN_PASSWORD
fi

# Ensure repository directory exists
mkdir -p "$REPO_STORAGE" 2>/dev/null || true
chmod 755 "$REPO_STORAGE" 2>/dev/null || true

# Start the application
echo "Starting MyGit v1.0.0..."
echo "  HTTP Port: $HTTP_PORT"
echo "  Repository Storage: $REPO_STORAGE"
echo "  Admin User: $ADMIN_USERNAME"

if [ -n "$ADMIN_PASSWORD" ]; then
    echo "  Admin Password: [SET]"
else
    echo "  Admin Password: [NOT SET - Web UI will be read-only]"
fi

exec /usr/local/bin/mygit