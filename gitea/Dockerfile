ARG BUILD_FROM
FROM $BUILD_FROM

ARG GITEA_VERSION=1.21
ARG BUILD_ARCH

WORKDIR /data

RUN apk add --no-cache curl gcompat git && \
     case "${BUILD_ARCH}" in \
         amd64)  ARCH="amd64" ;; \
         aarch64) ARCH="arm64" ;; \
         armv7)  ARCH="arm" ;; \
         i386)   ARCH="386" ;; \
         *)      ARCH="amd64" ;; \
     esac && \
      curl -fsSL "https://dl.gitea.com/gitea/${GITEA_VERSION}/gitea-${GITEA_VERSION}-linux-${ARCH}" -o /usr/local/bin/gitea && \
      chmod +x /usr/local/bin/gitea && \
      addgroup -g ${USER_GID:-1000} git && \
      adduser -D -u ${USER_UID:-1000} -G git -h /data/git -s /bin/sh git && \
      mkdir -p /data/gitea /data/git/repositories /data/git/custom && \
      chown -R git:git /data && \
      chmod -R 775 /data

COPY run.sh /
COPY s6-services/gitea/ /etc/services.d/gitea/
RUN chmod a+x /run.sh
