#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：管理运行期 conversation session，用于缓存、续连和运行态优化。
# 请求：PUT /api/agent-factory/v1/conversation/session/{conversation_id}
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用
require_env CONVERSATION_ID "set CONVERSATION_ID before managing session."  # 确保 CONVERSATION_ID 存在
ensure_agent_id  # 确保 AGENT_ID 存在

# 使用 jq 创建 JSON 对象。
jq -n \
  --arg action "${SESSION_ACTION:-get_info_or_create}" \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v1}" \
  '{
    action: $action,
    agent_id: $agent_id,
    agent_version: $agent_version
  }' \
  | af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/conversation/session/$CONVERSATION_ID" \
      -H "Content-Type: application/json" \
      --data-binary @- \
  | jq .
