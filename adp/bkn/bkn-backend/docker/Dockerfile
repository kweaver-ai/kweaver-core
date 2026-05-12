# Copyright The kweaver.ai Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.

########################### Stage 0 ########################
ARG BUILD_IMAGE=golang:1.25.10
ARG BASE_IMAGE=ubuntu:24.04

FROM ${BUILD_IMAGE} AS builder

COPY ./server /go/src

ARG SERVER_VERSION=0.1.0
ARG GOPROXY_URL="http://goproxy.cn,direct"

ENV GOPROXY=${GOPROXY_URL}

RUN set -ex; \
    cd /go/src; \
    go mod download -x; \
    export flags="-X 'bkn-backend/version.ServerVersion=${SERVER_VERSION}'"; \
    CGO_ENABLED=1 go build -ldflags "-s -w $flags" -o ./bin/bkn-backend-server;

########################### Stage 1 ########################
# Copy working directory to the actual release docker images
FROM ${BASE_IMAGE} AS prod

ARG UID=1001
ARG GID=1001

RUN groupadd -r -g $GID kweaver \
    && useradd -r -u $UID -g $GID kweaver

COPY --from=builder --chown=$UID:$GID /go/src/bin /opt/bkn-backend

ADD server/config/bkn-backend-config.yaml /opt/bkn-backend/config/
ADD server/locale /opt/bkn-backend/locale

# 指定工作目录
WORKDIR /opt/bkn-backend/

# Change user
USER $UID

# Expose port
EXPOSE 13014

# 指定启动命令
ENTRYPOINT [ "./bkn-backend-server" ]
