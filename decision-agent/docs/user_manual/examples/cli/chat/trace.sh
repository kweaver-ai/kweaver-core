#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询指定对话的 trace。后端未开放 trace 路由时会给出跳过说明。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在
skip_if_empty CONVERSATION_ID "set CONVERSATION_ID before reading trace."  # 检查CONVERSATION_ID是否为空

if ! kweaver agent trace "$AGENT_ID" "$CONVERSATION_ID" --pretty; then  # 尝试调用kweaver命令查询trace，如果失败则执行下面的语句
  echo "Skip trace output: current backend may not expose the trace route, or trace-ai service is unavailable."  # 输出跳过trace的说明信息
fi
