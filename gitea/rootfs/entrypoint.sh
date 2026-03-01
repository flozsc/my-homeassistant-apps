#!/bin/sh
set -e

echo "==> Gitea addon starting..."
echo "==> root_url='$root_url'"

CONFIG_FILE="/data/gitea/conf/app.ini"

if [ -f "$CONFIG_FILE" ]; then
    echo "==> Current ROOT_URL in config:"
    grep "^ROOT_URL" "$CONFIG_FILE" || echo "==> ROOT_URL not found in config"
    
    if [ -n "$root_url" ]; then
        echo "==> Setting ROOT_URL to: $root_url"
        sed -i "s|^ROOT_URL\s*=.*|ROOT_URL = ${root_url}|" "$CONFIG_FILE"
        echo "==> ROOT_URL after sed:"
        grep "^ROOT_URL" "$CONFIG_FILE" || echo "==> ROOT_URL not found after sed"
    else
        echo "==> root_url is empty, not modifying config"
    fi
else
    echo "==> No config file found"
fi

echo "==> Starting Gitea..."
exec su-exec git /usr/local/bin/gitea web --config /data/gitea/conf/app.ini
