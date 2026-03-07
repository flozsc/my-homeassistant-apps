# MyGit - Private Git Server for Home Assistant
# Single-stage build that works with Home Assistant's build system

FROM ghcr.io/home-assistant/amd64-base:latest

# Install build dependencies
RUN apk add --no-cache git openssh-server ca-certificates && \
    wget -q -O /tmp/go.tar.gz https://go.dev/dl/go1.26.1.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz && \
    export PATH=/usr/local/go/bin:$PATH

# Create git user and directories
RUN addgroup -S git && \
    adduser -S -G git git && \
    mkdir -p /data/repos && \
    chown git:git /data/repos && \
    chmod 755 /data/repos

# Build the application
WORKDIR /app
COPY . .
RUN /usr/local/go/bin/go build -o /usr/local/bin/mygit ./src/main.go

# Copy S6 service configuration
COPY rootfs/ / 

# Set permissions for S6 service files
RUN chmod +x /etc/services.d/mygit/run /etc/services.d/mygit/finish && \
    chown -R root:root /etc/services.d/mygit

USER git