#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：调用 Debug 对话。Debug 请求把 query 放在 input 内，适合配置页调试。
# 请求：POST /api/agent-factory/v1/app/{agent_key}/debug/completion
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
ensure_agent_id  # 确保 AGENT_ID 存在
require_env AGENT_KEY "set AGENT_KEY from a published Agent."  # 确保 AGENT_KEY 存在

# 发送 Debug 对话请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"'"$AGENT_ID"'","agent_version":"'"${AGENT_VERSION:-v0}"'","input":{"query":"'"${CHAT_MESSAGE:-测试 debug completion}"'","history":[],"custom_querys":{}},"stream":false}'  # 发送请求体
echo  # 输出空行
