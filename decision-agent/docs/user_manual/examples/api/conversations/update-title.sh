#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：更新对话标题。
# 请求：PUT /api/agent-factory/v1/app/{agent_key}/conversation/{conversation_id}
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用
require_env AGENT_KEY "set AGENT_KEY before updating conversation."  # 确保 AGENT_KEY 存在
require_env CONVERSATION_ID "set CONVERSATION_ID before updating conversation."  # 确保 CONVERSATION_ID 存在

# 发送更新标题请求。
af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -d '{"id":"'"$CONVERSATION_ID"'","title":"'"${CONVERSATION_TITLE:-示例对话标题}"'"}' \
  | jq .
