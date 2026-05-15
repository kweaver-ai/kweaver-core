#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：终止一次正在执行的对话 run。
# 请求：POST /api/agent-factory/v1/app/{agent_key}/chat/termination
# 可从流式响应或消息详情中取得 AGENT_RUN_ID / INTERRUPTED_ASSISTANT_MESSAGE_ID。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_env AGENT_KEY "set AGENT_KEY from a published Agent."  # 确保 AGENT_KEY 存在
require_env CONVERSATION_ID "set CONVERSATION_ID before terminating chat."  # 确保 CONVERSATION_ID 存在

# 发送终止对话请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/termination" \
  -H "Content-Type: application/json" \
  -d '{"conversation_id":"'"$CONVERSATION_ID"'","agent_run_id":"'"${AGENT_RUN_ID:-}"'","interrupted_assistant_message_id":"'"${INTERRUPTED_ASSISTANT_MESSAGE_ID:-}"'"}'  # 发送请求体
echo  # 输出空行
