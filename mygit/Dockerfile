ARG BUILD_FROM
FROM $BUILD_FROM

ARG GIT_VERSION=2.45.2

# Install dependencies
RUN apk add --no-cache \
    git \
    openssh-server \
    ca-certificates \
    && update-ca-certificates

# Create git user
RUN addgroup -S git && \
    adduser -S -G git git && \
    mkdir -p /data/repos && \
    chown git:git /data/repos && \
    chmod 755 /data/repos

# Copy application
WORKDIR /app
COPY . .

# Build application
RUN go build -o /usr/local/bin/mygit ./src/main.go && \
    chown git:git /usr/local/bin/mygit && \
    chmod 755 /usr/local/bin/mygit

# Copy startup script
COPY run-test.sh /run.sh
RUN chmod +x /run.sh && \
    chown git:git /run.sh

USER git
CMD ["/run.sh"]