# Multi-Language Executor Runtime Base
# Stable base image with Python, Go, Bash, and common toolchain dependencies.
# This image does not contain executor application code.

ARG BASE_IMAGE=sandbox-python-executor-base:python3.11-v1
FROM ${BASE_IMAGE}

ARG USE_MIRROR=false
ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG GO_VERSION=1.25.2
ARG GO_DOWNLOAD_BASE=https://go.dev/dl
ARG GO_DOWNLOAD_MIRROR=https://mirrors.ustc.edu.cn/golang
# docker buildx sets TARGETARCH from --platform for each target architecture.
ARG TARGETARCH

LABEL maintainer="Sandbox Team"
LABEL description="Sandbox multi-language executor runtime base without executor code"
LABEL version="1.0.0"

USER root

RUN if [ "$USE_MIRROR" = "true" ]; then \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list; \
    fi

RUN set -eux; \
    arch="${TARGETARCH:-amd64}"; \
    case "$arch" in \
        amd64) go_arch="amd64" ;; \
        arm64) go_arch="arm64" ;; \
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    go_download_base="$GO_DOWNLOAD_BASE"; \
    if [ "$USE_MIRROR" = "true" ]; then \
        go_download_base="$GO_DOWNLOAD_MIRROR"; \
    fi; \
    apt-get update; \
    apt-get install -y --no-install-recommends git; \
    curl -fsSL \
        --retry 5 \
        --retry-delay 3 \
        --retry-all-errors \
        --connect-timeout 20 \
        "${go_download_base%/}/go${GO_VERSION}.linux-${go_arch}.tar.gz" \
        -o /tmp/go.tgz; \
    rm -rf /usr/local/go; \
    tar -C /usr/local -xzf /tmp/go.tgz; \
    rm -f /tmp/go.tgz; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/local/go/bin:${PATH}" \
    GOCACHE="/workspace/.cache/go-build" \
    GOMODCACHE="/workspace/.cache/go-mod"

USER sandbox
