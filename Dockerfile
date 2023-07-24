# Copyright (c) Curt Bushko.
# SPDX-License-Identifier: MPL-2.0
ARG GOVERSION=1.18
FROM golang:${GOVERSION} AS builder
MAINTAINER Curt Bushko (https://github.com/curtbushko)
LABEL org.opencontainers.image.source=https://github.com/curtbushko/commit-status-action

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apt-get -qq update && apt-get -yqq install upx file

WORKDIR /src
COPY . .

RUN mkdir -p /bin && go build -o /bin/action .

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

FROM scratch
MAINTAINER Curt Bushko (https://github.com/curtbushko)

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder --chown=65534:0 /bin/action /action

USER nobody
ENTRYPOINT ["/action"]
