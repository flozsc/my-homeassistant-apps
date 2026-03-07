#!/usr/bin/with-contenv bashio

# MyGit Run Script - Reads config from HA Supervisor API

HTTP_PORT=$(bashio::config 'http_port')
ADMIN_USERNAME=$(bashio::config 'admin_username')
REPO_STORAGE=$(bashio::config 'repo_storage')

export HTTP_PORT ADMIN_USERNAME REPO_STORAGE

if bashio::config.exists 'admin_password'; then
    export ADMIN_PASSWORD=$(bashio::config 'admin_password')
fi

bashio::log.info "Starting MyGit..."
bashio::log.info "  HTTP Port: ${HTTP_PORT}"
bashio::log.info "  Admin User: ${ADMIN_USERNAME}"
bashio::log.info "  Repository Storage: ${REPO_STORAGE}"

exec /usr/local/bin/mygit
