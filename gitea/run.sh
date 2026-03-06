#!/usr/bin/with-contenv bashio

bashio::log.info "Starting Gitea..."

ROOT_URL=$(bashio::config 'root_url')
ADMIN_USERNAME=$(bashio::config 'admin_username')
ADMIN_EMAIL=$(bashio::config 'admin_email')
HTTP_PORT=$(bashio::config 'http_port')
SSH_PORT=$(bashio::config 'ssh_port')

mkdir -p /data/gitea
mkdir -p /data/git

export GITEA__server__ROOT_URL="${ROOT_URL}"
export GITEA__server__HTTP_PORT="${HTTP_PORT}"
export GITEA__server__SSH_PORT="${SSH_PORT}"
export GITEA__server__SSH_DOMAIN="$(bashio::host.hostname)"
export GITEA__database__DB_TYPE=sqlite3
export GITEA__database__PATH=/data/gitea/gitea.db
export GITEA__service__DISABLE_REGISTRATION=true
export GITEA__security__INSTALL_LOCK=true

export GITEA__admin__DEFAULT_ADMIN_USER="${ADMIN_USERNAME}"
export GITEA__admin__DEFAULT_ADMIN_EMAIL="${ADMIN_EMAIL}"

if bashio::config.exists 'admin_password'; then
    ADMIN_PASSWORD=$(bashio::config 'admin_password')
    export GITEA__admin__DEFAULT_ADMIN_PASSWORD="${ADMIN_PASSWORD}"
    bashio::log.info "  Admin Password: [SET]"
fi

bashio::log.info "Gitea configuration:"
bashio::log.info "  Root URL: ${ROOT_URL}"
bashio::log.info "  HTTP Port: ${HTTP_PORT}"
bashio::log.info "  SSH Port: ${SSH_PORT}"
bashio::log.info "  Admin User: ${ADMIN_USERNAME}"
bashio::log.info "  Registration: Disabled"

exec /usr/bin/entrypoint /bin/sh
