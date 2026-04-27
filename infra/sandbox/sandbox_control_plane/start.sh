#!/bin/bash
# Sandbox Control Plane 启动脚本

# 设置项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PYTHONPATH="$SCRIPT_DIR"

# 进入项目目录
cd "$SCRIPT_DIR"

# 清理端口 8000 上的旧进程
echo "检查端口 8000..."
OLD_PIDS=$(lsof -ti:8000 2>/dev/null)
if [ -n "$OLD_PIDS" ]; then
    echo "发现旧进程占用端口 8000: $OLD_PIDS"
    echo "正在终止旧进程..."
    lsof -ti:8000 | xargs kill -9 2>/dev/null
    sleep 1
    echo "旧进程已清理"
fi

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "警告: .env 文件不存在"
    echo "正在从 .env.example 创建 .env..."
    cp .env.example .env
    echo "请根据需要编辑 .env 文件"
fi

# 同步依赖
echo "正在同步依赖..."
if command -v uv &> /dev/null; then
    uv sync
else
    echo "uv 未安装，使用 pip 安装依赖..."
    pip install -e ".[dev]"
fi

# 启动服务
echo "正在启动 Sandbox Control Plane..."
echo "提示: 使用 Ctrl+C 停止服务"
echo ""
echo "服务将在以下地址可用:"
echo "  - HTTP: http://localhost:8000"
echo "  - API 文档: http://localhost:8000/docs"
echo "  - ReDoc: http://localhost:8000/redoc"
echo ""

# 使用 uv 或直接运行
if command -v uv &> /dev/null; then
    uv run uvicorn src.interfaces.rest.main:app --host 0.0.0.0 --port 8000 --reload
else
    uvicorn src.interfaces.rest.main:app --host 0.0.0.0 --port 8000 --reload
fi
