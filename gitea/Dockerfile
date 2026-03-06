ARG BUILD_FROM
FROM $BUILD_FROM

COPY run.sh /
RUN chmod a+x /run.sh

WORKDIR /data

ENTRYPOINT [ "/run.sh" ]
CMD [ ]
