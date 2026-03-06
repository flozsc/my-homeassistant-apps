ARG BUILD_FROM
FROM $BUILD_FROM

ARG GITEA_VERSION=1.21
ARG TARGETARCH

RUN apk add --no-cache curl gcompat git

WORKDIR /data

RUN case "${TARGETARCH}" in \
        amd64)  ARCH="amd64" ;; \
        arm64)  ARCH="arm64" ;; \
        arm)    ARCH="arm" ;; \
        386)    ARCH="386" ;; \
        *)      ARCH="amd64" ;; \
    esac && \
    curl -fsSL "https://dl.gitea.com/gitea/${GITEA_VERSION}/gitea-${GITEA_VERSION}-linux-${ARCH}" -o /usr/local/bin/gitea && \
    chmod +x /usr/local/bin/gitea && \
    addgroup -S git && \
    adduser -S git -G git && \
    mkdir -p /data/gitea /data/git/repositories /data/git/custom && \
    chown -R git:git /data

USER git

COPY run.sh /
RUN chmod a+x /run.sh

ENTRYPOINT [ "/run.sh" ]
CMD []
