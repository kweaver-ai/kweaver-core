# Python Executor Runtime Base
# Stable base image with Python and sandbox runtime system dependencies.
# This image does not contain executor application code.

ARG USE_MIRROR=false
ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG PYTHON_VERSION=3.11
ARG PYTHON_IMAGE=python:${PYTHON_VERSION}-slim

FROM ${PYTHON_IMAGE}

ARG USE_MIRROR=false
ARG APT_MIRROR=mirrors.ustc.edu.cn

LABEL maintainer="Sandbox Team"
LABEL description="Sandbox Python executor runtime base without executor code"
LABEL version="1.0.0"

RUN if [ "$USE_MIRROR" = "true" ]; then \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list; \
    fi

RUN apt-get update && apt-get install -y --no-install-recommends \
    bubblewrap \
    s3fs \
    gosu \
    curl \
    ca-certificates \
    bash \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

RUN groupadd -g 1000 sandbox && \
    useradd -m -u 1000 -g sandbox sandbox

WORKDIR /app

RUN mkdir -p /workspace /opt/sandbox-venv && \
    chown -R sandbox:sandbox /app /workspace /opt/sandbox-venv

ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1
ENV WORKSPACE_PATH=/workspace

USER sandbox
