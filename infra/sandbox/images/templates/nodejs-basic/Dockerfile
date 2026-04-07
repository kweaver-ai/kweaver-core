# Node.js Basic Template
# Minimal Node.js 20 environment for JavaScript/TypeScript execution
# Based on sandbox-executor-base which contains the executor daemon

ARG BASE_IMAGE=sandbox-executor-base:latest
FROM ${BASE_IMAGE}

# Metadata
LABEL template.name="nodejs-basic"
LABEL template.description="Node.js 20 LTS environment"
LABEL template.version="1.0.0"

# Switch to root to install Node.js
USER root

# Install Node.js 20 LTS using official installer
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    node --version && \
    npm --version && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Switch back to sandbox user
USER sandbox

# CMD is inherited from the base image (executor daemon)
