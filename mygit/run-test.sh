#!/bin/sh

# Set default values
HTTP_PORT=${HTTP_PORT:-3000}
REPO_STORAGE=${REPO_STORAGE:-/data/repos}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}

# Ensure repository directory exists
if [ ! -d "$REPO_STORAGE" ]; then
    mkdir -p "$REPO_STORAGE" 2>/dev/null || true
fi
chmod 755 "$REPO_STORAGE" 2>/dev/null || true

echo "Starting MyGit v1.0.0..."
echo "  HTTP Port: $HTTP_PORT"
echo "  Repository Storage: $REPO_STORAGE"
echo "  Admin User: $ADMIN_USERNAME"

exec /usr/local/bin/mygit