#!/bin/sh
set -e

echo "==> Gitea addon starting..."
echo "==> ROOT_URL='$ROOT_URL'"

CONFIG_FILE="/data/gitea/conf/app.ini"

if [ -f "$CONFIG_FILE" ]; then
    echo "==> Current ROOT_URL in config:"
    grep "^ROOT_URL" "$CONFIG_FILE" || echo "==> ROOT_URL not found in config"
    
    if [ -n "$ROOT_URL" ]; then
        echo "==> Setting ROOT_URL to: $ROOT_URL"
        sed -i "s|^ROOT_URL\s*=.*|ROOT_URL = ${ROOT_URL}|" "$CONFIG_FILE"
        echo "==> ROOT_URL after sed:"
        grep "^ROOT_URL" "$CONFIG_FILE" || echo "==> ROOT_URL not found after sed"
    else
        echo "==> ROOT_URL is empty, not modifying config"
    fi
else
    echo "==> No config file found"
fi

echo "==> Starting Gitea..."
exec su-exec git /usr/local/bin/gitea web --config /data/gitea/conf/app.ini
