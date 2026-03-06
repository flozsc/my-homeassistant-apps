#!/usr/bin/with-contenv bashio

bashio::log.info "Starting Gitea..."

ROOT_URL=$(bashio::config 'root_url')
ADMIN_USERNAME=$(bashio::config 'admin_username')
ADMIN_EMAIL=$(bashio::config 'admin_email')
HTTP_PORT=$(bashio::config 'http_port')
SSH_PORT=$(bashio::config 'ssh_port')

mkdir -p /data/gitea
mkdir -p /data/git

cat > /data/gitea/app.ini << EOF
APP_NAME = Gitea
RUN_USER = git
RUN_MODE = prod

[repository]
ROOT = /data/git/repositories

[database]
DB_TYPE = sqlite3
PATH = /data/gitea/gitea.db

[server]
DOMAIN = $(hostname -f)
HTTP_PORT = ${HTTP_PORT}
ROOT_URL = ${ROOT_URL}
SSH_PORT = ${SSH_PORT}
SSH_DOMAIN = $(hostname -f)
START_SSH_SERVER = true

[service]
DISABLE_REGISTRATION = true

[security]
INSTALL_LOCK = true
SECRET_KEY =

[admin]
DEFAULT_ADMIN_USER = ${ADMIN_USERNAME}
DEFAULT_ADMIN_EMAIL = ${ADMIN_EMAIL}
EOF

if bashio::config.exists 'admin_password'; then
    ADMIN_PASSWORD=$(bashio::config 'admin_password')
    sed -i "/DEFAULT_ADMIN_EMAIL/a DEFAULT_ADMIN_PASSWORD = ${ADMIN_PASSWORD}" /data/gitea/app.ini
    bashio::log.info "  Admin Password: [SET]"
fi

bashio::log.info "Gitea configuration:"
bashio::log.info "  Root URL: ${ROOT_URL}"
bashio::log.info "  HTTP Port: ${HTTP_PORT}"
bashio::log.info "  SSH Port: ${SSH_PORT}"
bashio::log.info "  Admin User: ${ADMIN_USERNAME}"
bashio::log.info "  Registration: Disabled"

exec /app/gitea/gitea web --config /data/gitea/app.ini --custom-path /data/git/custom --work-path /data/git
