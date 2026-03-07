#!/bin/sh

# MyGit Run Script for Home Assistant
# Works with HA's built-in S6 overlay when init: true

# Set defaults (Home Assistant will override via environment variables)
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

# Directory setup
mkdir -p "$REPO_STORAGE" 2>/dev/null || true
chmod 755 "$REPO_STORAGE" 2>/dev/null || true

# Log startup
echo "[MyGit] Starting v0.0.4 with Home Assistant S6 overlay"
echo "[MyGit] Configuration: Port=$HTTP_PORT, User=$ADMIN_USERNAME, Storage=$REPO_STORAGE"

# Execute the application
exec /usr/local/bin/mygit