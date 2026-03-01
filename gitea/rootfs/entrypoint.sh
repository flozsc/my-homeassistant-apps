#!/bin/sh
set -e

echo "==> Gitea addon starting..."
echo "==> ENV vars:"
env | grep GITEA || echo "==> No GITEA env vars found"

echo "==> Looking for gitea binary..."
which gitea || ls -la /usr/local/bin/ || true

CONFIG_FILE="/data/gitea/conf/app.ini"

if [ -f "$CONFIG_FILE" ]; then
    echo "==> Found config at $CONFIG_FILE"
    
    if [ -n "$GITEA_SSH_PORT" ]; then
        echo "==> SSH_PORT = $GITEA_SSH_PORT"
        if [ "$GITEA_SSH_PORT" = "0" ]; then
            sed -i "s/^SSH_PORT = .*/SSH_PORT = -1/" "$CONFIG_FILE"
        else
            sed -i "s/^SSH_PORT = .*/SSH_PORT = $GITEA_SSH_PORT/" "$CONFIG_FILE"
        fi
    fi

    if [ -n "$GITEA_ROOT_URL" ]; then
        echo "==> Setting ROOT_URL to: $GITEA_ROOT_URL"
        sed -i "s|^ROOT_URL\s*=.*|ROOT_URL = ${GITEA_ROOT_URL}|" "$CONFIG_FILE"
    fi
else
    echo "==> No config file found"
fi

exec "$(which gitea)" web --config /data/gitea/conf/app.ini
