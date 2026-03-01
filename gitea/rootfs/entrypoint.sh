#!/bin/sh
set -e

CONFIG_FILE="/data/gitea/conf/app.ini"
OPTIONS_FILE="/data/options.json"

if [ -f "$OPTIONS_FILE" ]; then
    ROOT_URL=$(grep -o '"root_url"[[:space:]]*:[[:space:]]*"[^"]*"' "$OPTIONS_FILE" | sed 's/.*"root_url"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    SSH_PORT=$(grep -o '"ssh_port"[[:space:]]*:[[:space:]]*[0-9]*' "$OPTIONS_FILE" | sed 's/.*"ssh_port"[[:space:]]*:[[:space:]]*\([0-9]*\)/\1/')
    ADMIN_PASSWORD=$(grep -o '"admin_password"[[:space:]]*:[[:space:]]*"[^"]*"' "$OPTIONS_FILE" | sed 's/.*"admin_password"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    
    if [ -n "$ADMIN_PASSWORD" ]; then
        export GITEA_ADMIN_PASSWORD="$ADMIN_PASSWORD"
        export GITEA_ADMIN_USERNAME="gitea_admin"
    fi
fi

if [ -f "$CONFIG_FILE" ]; then
    if [ -n "$ROOT_URL" ]; then
        sed -i "s|^ROOT_URL\s*=.*|ROOT_URL = ${ROOT_URL}|" "$CONFIG_FILE"
    fi
    
    if [ -n "$SSH_PORT" ]; then
        if [ "$SSH_PORT" = "0" ]; then
            sed -i "s/^SSH_PORT = .*/SSH_PORT = -1/" "$CONFIG_FILE"
        else
            sed -i "s/^SSH_PORT = .*/SSH_PORT = $SSH_PORT/" "$CONFIG_FILE"
        fi
    fi
fi

exec su-exec git /usr/local/bin/gitea web --config /data/gitea/conf/app.ini
