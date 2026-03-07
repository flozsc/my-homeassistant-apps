#!/bin/sh

# Set default values
HTTP_PORT=${HTTP_PORT:-3000}
REPO_STORAGE=${REPO_STORAGE:-/data/repos}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}

# If admin password is set, use it
if [ -n "${ADMIN_PASSWORD}" ]; then
    export ADMIN_PASSWORD
fi

# Ensure repository directory exists
mkdir -p "$REPO_STORAGE" 2>/dev/null || true
chmod 755 "$REPO_STORAGE" 2>/dev/null || true

# Start the application
exec /usr/local/bin/mygit