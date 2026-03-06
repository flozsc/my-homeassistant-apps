ARG BUILD_FROM
FROM $BUILD_FROM

RUN apk add --no-cache curl gcompat

WORKDIR /data

COPY run.sh /
RUN chmod a+x /run.sh

ENV USER_UID=1000
ENV USER_GID=1000
ENV GITEA__database__DB_TYPE=sqlite3
ENV GITEA__database__PATH=/data/gitea/gitea.db
ENV GITEA__service__DISABLE_REGISTRATION=true
ENV GITEA__security__INSTALL_LOCK=true

ENTRYPOINT [ "/run.sh" ]
CMD []
