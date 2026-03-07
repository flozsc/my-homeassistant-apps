#!/bin/sh

# MyGit Run Script - Simple and robust approach
# Works with Home Assistant's built-in S6 overlay (init: true)
# Also suitable for local development with environment variables

# Set defaults (environment variables override these)
HTTP_PORT=${HTTP_PORT:-3000}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
REPO_STORAGE=${REPO_STORAGE:-/data/repos}

# Export for the application
export HTTP_PORT
export ADMIN_USERNAME
export REPO_STORAGE

# Password handling - only set if explicitly configured
if [ -n "${ADMIN_PASSWORD}" ]; then
    export ADMIN_PASSWORD
fi

# Directory setup with error handling
mkdir -p "$REPO_STORAGE" 2>/dev/null || true
chmod 755 "$REPO_STORAGE" 2>/dev/null || true

# Start application
echo "Starting MyGit v0.0.4..."
echo "  Port: $HTTP_PORT"
echo "  User: $ADMIN_USERNAME"
echo "  Storage: $REPO_STORAGE"

# Execute the application
exec /usr/local/bin/mygit