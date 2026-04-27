# Sandbox Executor Base Image
# This image serves as the base for all template images
# Contains the executor daemon and Bubblewrap for secure code execution

# Mirror configuration for special network environments
ARG USE_MIRROR=false
ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG PIP_MIRROR=mirrors.ustc.edu.cn
ARG PYTHON_VERSION=3.11
ARG PYTHON_IMAGE=python:${PYTHON_VERSION}-slim

FROM ${PYTHON_IMAGE} AS builder

# Re-declare ARG after FROM (ARGs don't persist across FROM)
ARG USE_MIRROR=false
ARG APT_MIRROR=mirrors.ustc.edu.cn
ARG PIP_MIRROR=mirrors.ustc.edu.cn

# Metadata
LABEL maintainer="Sandbox Team"
LABEL description="Sandbox Executor Base - Daemon + Bubblewrap Isolation"
LABEL version="1.0.0"

# Configure APT mirror if needed
RUN if [ "$USE_MIRROR" = "true" ]; then \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list; \
    fi

# Install build and runtime dependencies required during dependency resolution
RUN apt-get update && apt-get install -y \
    bubblewrap \
    s3fs \
    gosu \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install uv (fast Python package installer)
RUN if [ "$USE_MIRROR" = "true" ]; then \
      pip install --no-cache-dir -i "https://$PIP_MIRROR/pypi/simple" uv; \
    else \
      pip install --no-cache-dir uv; \
    fi

# Create non-privileged user
# The executor runs as UID:GID 1000:1000 for security
RUN groupadd -g 1000 sandbox && \
    useradd -m -u 1000 -g sandbox sandbox

WORKDIR /app

# Copy dependency files first (better layer caching)
COPY runtime/executor/pyproject.toml runtime/executor/uv.lock ./

# Install Python dependencies using uv (faster, uses lock file)
RUN if [ "$USE_MIRROR" = "true" ]; then \
      export UV_INDEX_URL="https://$PIP_MIRROR/pypi/simple"; \
    fi && \
    uv sync --frozen --no-dev

FROM ${PYTHON_IMAGE} AS runtime

ARG USE_MIRROR=false
ARG APT_MIRROR=mirrors.ustc.edu.cn

LABEL maintainer="Sandbox Team"
LABEL description="Sandbox Executor Base - Daemon + Bubblewrap Isolation"
LABEL version="1.0.0"

# Configure APT mirror if needed
RUN if [ "$USE_MIRROR" = "true" ]; then \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
      sed -i "s|deb.debian.org|$APT_MIRROR|g" /etc/apt/sources.list; \
    fi

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    bubblewrap \
    s3fs \
    gosu \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create non-privileged user
# The executor runs as UID:GID 1000:1000 for security
RUN groupadd -g 1000 sandbox && \
    useradd -m -u 1000 -g sandbox sandbox

WORKDIR /app

# Copy Python virtual environment from builder stage
COPY --from=builder /app/.venv ./.venv

# Copy executor code
COPY runtime/executor/ ./executor/

# Create workspace and venv directories
# /workspace will be mounted as a volume for S3/ephemeral storage
# /opt/sandbox-venv is for runtime dependency installation
RUN mkdir -p /workspace /opt/sandbox-venv && \
    chown -R sandbox:sandbox /app /workspace /opt/sandbox-venv

# Switch to non-privileged user
USER sandbox

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1
ENV WORKSPACE_PATH=/workspace
ENV PYTHONPATH=/app/.venv/lib/python3.11/site-packages:/app
ENV PATH="/app/.venv/bin:$PATH"

# Expose executor API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Start the executor daemon
# The executor runs as PID 1 and manages its lifecycle
# Note: When dependencies are specified, control plane will inject
# a custom entrypoint script to install them before starting the executor
CMD [".venv/bin/python", "-m", "executor.interfaces.http.rest"]
