#!/bin/bash
# Sandbox Control Plane 停止脚本

echo "正在停止 Sandbox Control Plane..."

# 查找并终止占用端口 8000 的进程
PIDS=$(lsof -ti:8000 2>/dev/null)

if [ -z "$PIDS" ]; then
    echo "未发现运行在端口 8000 的进程"
    exit 0
fi

echo "发现以下进程:"
ps -p $PIDS -o pid,ppid,cmd 2>/dev/null || echo "PID: $PIDS"

echo ""
echo "正在终止进程..."

# 尝试优雅终止
for PID in $PIDS; do
    echo "发送 SIGTERM 到进程 $PID..."
    kill -TERM $PID 2>/dev/null
done

# 等待进程结束
echo "等待进程结束..."
for i in {1..10}; do
    sleep 1
    REMAINING=$(lsof -ti:8000 2>/dev/null)
    if [ -z "$REMAINING" ]; then
        echo "所有进程已优雅终止"
        exit 0
    fi
    echo "等待中... ($i/10)"
done

# 强制终止剩余进程
REMAINING=$(lsof -ti:8000 2>/dev/null)
if [ -n "$REMAINING" ]; then
    echo "强制终止剩余进程..."
    lsof -ti:8000 | xargs kill -9 2>/dev/null
    sleep 1
fi

# 验证
REMAINING=$(lsof -ti:8000 2>/dev/null)
if [ -z "$REMAINING" ]; then
    echo "Sandbox Control Plane 已停止"
else
    echo "警告: 部分进程可能仍在运行"
    echo "剩余进程: $REMAINING"
    exit 1
fi
