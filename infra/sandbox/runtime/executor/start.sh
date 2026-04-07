#!/bin/bash
# Runtime Executor 启动脚本

# 设置 PYTHONPATH
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PYTHONPATH="$(dirname "$SCRIPT_DIR")"

# 进入项目目录
cd "$(dirname "$0")"

# 清理端口 8080 上的旧进程
echo "检查端口 8080..."
OLD_PIDS=$(lsof -ti:8080 2>/dev/null)
if [ -n "$OLD_PIDS" ]; then
    echo "发现旧进程占用端口 8080: $OLD_PIDS"
    echo "正在终止旧进程..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null
    sleep 1
    echo "旧进程已清理"
fi

# 同步依赖
echo "正在同步依赖..."
uv sync

# 启动服务
echo "正在启动服务..."
echo "提示: 使用 Ctrl+C 停止服务"
echo ""
uv run uvicorn executor.interfaces.http.rest:app --host 0.0.0.0 --port 8080 --reload
