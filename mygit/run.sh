#!/bin/sh

# MyGit Run Script - Works with or without HA Supervisor

# Use environment variables with sensible defaults
HTTP_PORT=${HTTP_PORT:-3000}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
REPO_STORAGE=${HTTP_REPO_STORAGE:-/data/repos}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-admin}

export HTTP_PORT ADMIN_USERNAME REPO_STORAGE ADMIN_PASSWORD

# Create storage directory
mkdir -p "$REPO_STORAGE" 2>/dev/null || true

echo "Starting MyGit v0.1.2 on port $HTTP_PORT..."
echo "Storage: $REPO_STORAGE"
echo "User: $ADMIN_USERNAME"

# Execute the application
exec /usr/local/bin/mygit
