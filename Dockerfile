FROM ubuntu:24.04

EXPOSE 1080

HEALTHCHECK --retries=10 CMD [ "/vignet", "-url", "http://127.0.0.1:1080/health" ]
ENTRYPOINT [ "/vignet" ]

VOLUME /tmp

ARG VERSION
ENV VERSION=${VERSION}

ARG GIT_SHA
ENV GIT_SHA=${GIT_SHA}

ARG TARGETOS
ARG TARGETARCH

USER 65534

COPY ffmpeg/${TARGETOS}/${TARGETARCH}/bin/ffmpeg /usr/bin/ffmpeg
COPY ffmpeg/${TARGETOS}/${TARGETARCH}/bin/ffprobe /usr/bin/ffprobe

COPY wait_${TARGETOS}_${TARGETARCH} /wait

COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY release/vignet_${TARGETOS}_${TARGETARCH} /vignet
