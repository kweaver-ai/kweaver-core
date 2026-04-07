# Copyright The kweaver.ai Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.

########################### Stage 0 ########################
ARG BUILD_IMAGE=golang:1.24.11
ARG BASE_IMAGE=ubuntu:24.04

FROM ${BUILD_IMAGE} AS builder

COPY ./server /go/src

ARG SERVER_VERSION=0.1.0
ARG GOPROXY_URL="http://goproxy.cn,direct"

ENV GOPROXY=${GOPROXY_URL}

RUN set -ex; \
    cd /go/src; \
    go mod download -x; \
    CGO_ENABLED=1 go build -ldflags "-s -w" -o ./bin/vega-backend-server;

########################### Stage 1 ########################
# Copy working directory to the actual release docker images
FROM ${BASE_IMAGE} AS prod

ARG UID=1001
ARG GID=1001

RUN groupadd -r -g $GID adp \
    && useradd -r -u $UID -g $GID adp

COPY --from=builder --chown=$UID:$GID /go/src/bin /opt/vega-backend

ADD server/config/vega-backend-config.yaml /opt/vega-backend/config/
ADD server/locale /opt/vega-backend/locale

# 指定工作目录
WORKDIR /opt/vega-backend/

# Change user
USER $UID

# Expose port
EXPOSE 13014

# 指定启动命令
ENTRYPOINT [ "./vega-backend-server" ]
