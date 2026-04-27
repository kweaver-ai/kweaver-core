# Agent Backend 统一Dockerfile
# 多阶段构建：构建3个可执行文件，使用supervisor管理进程

ARG buildImage=python:3.11-slim
ARG baseImage=ubuntu:noble
ARG goBuildImage=golang:1.24-alpine

# ==================== 构建阶段：agent-factory (Go) ====================
FROM ${goBuildImage} AS builder-factory

ARG ARCH_TYPE=amd64
ARG GOPROXY=https://goproxy.cn,direct

WORKDIR /build

# 复制完整 agent-factory 源代码
COPY agent-backend/agent-factory/ /build/

# 构建 agent-factory Go 二进制文件
RUN set -ex; \
    cd /build && \
    export GOPROXY=${GOPROXY} && \
    go mod download && \
    if [ "${ARCH_TYPE}" = "arm64" ]; then \
    go build -ldflags '-s -w' -o agent-factory ./main.go; \
    else \
    go build -ldflags '-s -w -checklinkname=0' -o agent-factory ./main.go; \
    fi && \
    chmod +x agent-factory

# ==================== 构建阶段：agent-executor (Python/PyInstaller) ====================
FROM ${buildImage} AS builder-executor

ARG ARCH_TYPE=amd64

WORKDIR /build

# 安装系统依赖和构建工具
RUN set -ex; \
    apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    make \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*; \
    pip install uv pyinstaller; \
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> /root/.bashrc

# 复制 agent-executor 源代码
COPY agent-backend/agent-executor/agent-executor.spec /build/agent-executor.spec
COPY agent-backend/agent-executor/pyproject.toml /build/pyproject.toml
COPY agent-backend/agent-executor/app /build/app
COPY agent-backend/agent-executor/main.py /build/main.py
COPY agent-backend/agent-executor/data_migrations /build/data_migrations
COPY agent-backend/agent-executor/hooks /build/hooks
# COPY agent-backend/agent-executor/mq-sdk-lib /build/mq-sdk-lib

# 安装依赖并构建 agent-executor
RUN set -ex; \
    cd /build; \
    export PATH="/root/.local/bin:$PATH"; \
    export UV_HTTP_TIMEOUT=300; \
    # sed -i 's|git+ssh://devops.aishu.cn:22/AISHUDevOps/One-Architecture/_git/proton-mq-python@1.7.4|git+https://github.com/kweaver-ai/proton-mq-sdk-py@release/1.9.0|' pyproject.toml; \
    # sed -i 's|git+ssh://devops.aishu.cn:22/AISHUDevOps/One-Architecture/_git/TelemetrySDK-Python@2.6.1|git+https://github.com/kweaver-ai/TelemetrySDK-Python@2.4.2|' pyproject.toml; \
    uv sync; \
    uv run pyinstaller agent-executor.spec

# ==================== 构建阶段：agent-memory (Python/PyInstaller) ====================
FROM ${buildImage} AS builder-memory

WORKDIR /build

# 安装系统依赖和构建工具
RUN set -ex; \
    apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    make \
    binutils \
    && rm -rf /var/lib/apt/lists/*; \
    pip install  pyinstaller

# 复制 agent-memory 源代码
COPY agent-backend/agent-memory/agent-memory.spec /build/agent-memory.spec
COPY agent-backend/agent-memory/requirements.txt /build/requirements.txt
COPY agent-backend/agent-memory/mem0 /build/mem0
COPY agent-backend/agent-memory/src /build/src

# 安装依赖并构建 agent-memory
RUN set -ex; \
    cd /build; \
    python3 -m pip install -i  https://mirrors.aliyun.com/pypi/simple  -r requirements.txt; \
    pyinstaller agent-memory.spec

# ==================== 生产阶段：supervisor 管理三个服务 ====================
FROM ${baseImage} AS prod

ARG ARCH_TYPE=amd64
ARG UID=1001
ARG GID=1001

# 安装 supervisor
RUN apt-get update && apt-get install -y \
    supervisor \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 创建非 root 用户
RUN groupadd -r -g $GID aishu && \
    useradd -r -u $UID -g $GID -d /home/aishu -m aishu && \
    mkdir -p /app/agent-factory /app/agent-executor /app/agent-memory /var/log/supervisor

WORKDIR /app

# 设置 mem0 数据目录，避免在 /home/aishu 创建目录的权限问题
ENV MEM0_DIR=/app/agent-memory

# 从构建阶段复制已构建的二进制文件
COPY --from=builder-factory --chown=$UID:$GID /build/agent-factory /app/agent-factory/agent-factory
COPY --from=builder-factory --chown=$UID:$GID /build/locale /app/agent-factory/locale
COPY --from=builder-executor --chown=$UID:$GID /build/dist/agent-executor /app/agent-executor
COPY --from=builder-memory --chown=$UID:$GID /build/dist/agent-memory /app/agent-memory

# # 处理 mq-sdk-lib 动态库
# COPY --from=builder-executor /build/mq-sdk-lib /tmp/mq-sdk-lib
# RUN if [ "$ARCH_TYPE" = "amd64" ]; then \
#     mv /tmp/mq-sdk-lib/x86/* /usr/lib/; \
#     elif [ "$ARCH_TYPE" = "arm64" ]; then \
#     mv /tmp/mq-sdk-lib/arm/* /usr/lib/; \
#     fi; \
#     rm -rf /tmp/mq-sdk-lib

# 复制 supervisor 配置文件
COPY --chown=$UID:$GID agent-backend/deploy/docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# 设置权限
RUN chown -R $UID:$GID /app /var/log/supervisor

# 切换到非 root 用户
USER $UID

# 暴露端口
# agent-factory: 13020
# agent-executor: 30778
# agent-memory: 30790
EXPOSE 13020 30778 30790

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:13020/health/ready && \
    curl -f http://localhost:30778/health/ready || exit 1

# 使用 supervisor 启动所有服务
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf", "-n"]
