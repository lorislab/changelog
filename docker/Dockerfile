FROM debian:10.7-slim AS builder
ARG VERSION=1.0.0

ENV FILENAME=changelog_${VERSION}_Linux_x86_64.tar.gz

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates

RUN curl https://github.com/lorislab/changelog/releases/download/${VERSION}/${FILENAME} -O -J -L && \
    tar xfz $FILENAME changelog && \
    chmod +x changelog

FROM debian:10.7-slim

LABEL org.opencontainers.image.source https://github.com/lorislab/changelog

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder changelog /opt/changelog

ENTRYPOINT ["/opt/changelog"]