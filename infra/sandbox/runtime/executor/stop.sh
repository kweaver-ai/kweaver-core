#!/bin/bash
# Runtime Executor 停止脚本

echo "正在停止 executor 服务..."

# 查找并终止占用端口 8080 的所有进程
PIDS=$(lsof -ti:8080 2>/dev/null)

if [ -z "$PIDS" ]; then
    echo "没有发现运行中的服务（端口 8080 未被占用）"
    exit 0
fi

echo "发现进程: $PIDS"
echo "正在终止..."

# 先尝试优雅退出（SIGTERM）
lsof -ti:8080 | xargs kill -15 2>/dev/null

# 等待 2 秒
sleep 2

# 检查是否还有进程
REMAINING=$(lsof -ti:8080 2>/dev/null)
if [ -n "$REMAINING" ]; then
    echo "进程未响应，强制终止（SIGKILL）..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null
    sleep 1
fi

# 最终检查
FINAL=$(lsof -ti:8080 2>/dev/null)
if [ -z "$FINAL" ]; then
    echo "✓ 服务已停止"
else
    echo "✗ 无法停止进程: $FINAL"
    exit 1
fi
