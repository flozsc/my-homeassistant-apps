ARG BUILD_FROM
FROM $BUILD_FROM

WORKDIR /data

COPY run.sh /
RUN chmod a+x /run.sh

ENTRYPOINT [ "/run.sh" ]
