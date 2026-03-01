#!/bin/sh
set -e

echo "==> Gitea addon starting..."

CONFIG_FILE="/data/gitea/conf/app.ini"
OPTIONS_FILE="/data/options.json"

if [ -f "$OPTIONS_FILE" ]; then
    echo "==> Reading options from $OPTIONS_FILE"
    ROOT_URL=$(grep -o '"root_url"[[:space:]]*:[[:space:]]*"[^"]*"' "$OPTIONS_FILE" | sed 's/.*"root_url"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    SSH_PORT=$(grep -o '"ssh_port"[[:space:]]*:[[:space:]]*[0-9]*' "$OPTIONS_FILE" | sed 's/.*"ssh_port"[[:space:]]*:[[:space:]]*\([0-9]*\)/\1/')
    echo "==> ROOT_URL from options: '$ROOT_URL'"
    echo "==> SSH_PORT from options: '$SSH_PORT'"
fi

if [ -f "$CONFIG_FILE" ]; then
    echo "==> Current ROOT_URL in config:"
    grep "^ROOT_URL" "$CONFIG_FILE" || echo "==> ROOT_URL not found in config"
    
    if [ -n "$ROOT_URL" ]; then
        echo "==> Setting ROOT_URL to: $ROOT_URL"
        sed -i "s|^ROOT_URL\s*=.*|ROOT_URL = ${ROOT_URL}|" "$CONFIG_FILE"
    fi
    
    if [ -n "$SSH_PORT" ]; then
        if [ "$SSH_PORT" = "0" ]; then
            sed -i "s/^SSH_PORT = .*/SSH_PORT = -1/" "$CONFIG_FILE"
        else
            sed -i "s/^SSH_PORT = .*/SSH_PORT = $SSH_PORT/" "$CONFIG_FILE"
        fi
    fi
else
    echo "==> No config file found"
fi

echo "==> Starting Gitea..."
exec su-exec git /usr/local/bin/gitea web --config /data/gitea/conf/app.ini
