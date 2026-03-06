#!/bin/sh
set -e

CONFIG_FILE="/data/gitea/conf/app.ini"
OPTIONS_FILE="/data/options.json"

mkdir -p /data/gitea/conf /data/gitea/log /data/gitea/data /data/gitea/data/repositories /data/gitea/data/tmp /data/gitea/data/jwt
chown -R git:git /data

HOSTNAME="localhost"
HTTP_PORT="3000"
SSH_PORT="2222"
ROOT_URL=""
ADMIN_PASSWORD=""

if [ -f "$OPTIONS_FILE" ]; then
    HOSTNAME=$(grep -o '"hostname"[[:space:]]*:[[:space:]]*"[^"]*"' "$OPTIONS_FILE" | sed 's/.*"hostname"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    HTTP_PORT=$(grep -o '"http_port"[[:space:]]*:[[:space:]]*[^,}]*' "$OPTIONS_FILE" | sed 's/.*"http_port"[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/')
    SSH_PORT=$(grep -o '"ssh_port"[[:space:]]*:[[:space:]]*[0-9]*' "$OPTIONS_FILE" | sed 's/.*"ssh_port"[[:space:]]*:[[:space:]]*\([0-9]*\)/\1/')
    ROOT_URL=$(grep -o '"root_url"[[:space:]]*:[[:space:]]*"[^"]*"' "$OPTIONS_FILE" | sed 's/.*"root_url"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
    ADMIN_PASSWORD=$(grep -o '"admin_password"[[:space:]]*:[[:space:]]*"[^"]*"' "$OPTIONS_FILE" | sed 's/.*"admin_password"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
fi

[ -z "$HOSTNAME" ] && HOSTNAME="localhost"
[ -z "$HTTP_PORT" ] && HTTP_PORT="3000"
[ -z "$SSH_PORT" ] && SSH_PORT="2222"

if [ -n "$ROOT_URL" ]; then
    ROOT_URL_LINE="ROOT_URL = $ROOT_URL"
else
    ROOT_URL_LINE="ROOT_URL = http://${HOSTNAME}:${HTTP_PORT}/"
fi

cat > "$CONFIG_FILE" << EOF
APP_NAME = Gitea: Git with a cup of tea
RUN_MODE = prod
RUN_USER = git
WORK_PATH = /data/gitea

[server]
PROTOCOL = http
DOMAIN = $HOSTNAME
HTTP_PORT = $HTTP_PORT
$ROOT_URL_LINE
DISABLE_SSH = false
SSH_PORT = $SSH_PORT
START_SSH_SERVER = true
LANDING_PAGE = home
APP_DATA_PATH = /data/gitea/data

[security]
INSTALL_LOCK = true
SECRET_KEY = 
INTERNAL_TOKEN = 
JWT_SECRET = 

[service]
DISABLE_REGISTRATION = false
REQUIRE_SIGNIN_VIEW = false
REGISTER_EMAIL_CONFIRM = false
ENABLE_NOTIFY_MAIL = false
DEFAULT_KEEP_EMAIL_PRIVATE = false
DEFAULT_ALLOW_CREATE_ORGANIZATION = true
DEFAULT_ENABLE_TIMETRACKING = true
NO_REPLY_ADDRESS = noreply.localhost

[oauth2]
ENABLE = false

[mailer]
ENABLED = false

[session]
PROVIDER = file

[log]
MODE = console, file
LEVEL = Info
ROOT_PATH = /data/gitea/log

[repository]
ROOT = /data/gitea/data/repositories
EOF

if [ -n "$ADMIN_PASSWORD" ]; then
    export GITEA_ADMIN_PASSWORD="$ADMIN_PASSWORD"
    export GITEA_ADMIN_USERNAME="gitea_admin"
    export GITEA_ADMIN_EMAIL="admin@localhost"
fi

export GITEA_WORK_DIR=/data/gitea
export GITEA_CUSTOM=/data/gitea

mkdir -p /data/gitea/{conf,log,data,data/repositories,data/tmp,data/jwt,data/lfs,data/attachments}
chown -R git:git /data

cd /data/gitea

chown git:git "$CONFIG_FILE"

exec su-exec git /usr/local/bin/gitea web --config /data/gitea/conf/app.ini
