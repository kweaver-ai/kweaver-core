#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：恢复读取断线前仍在运行的流式对话。
# 注意：这是流式续连，不等同于人工干预后的 resume_interrupt_info。
# 请求：POST /api/agent-factory/v1/app/{agent_key}/chat/resume
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_env AGENT_KEY "set AGENT_KEY from a published Agent."  # 确保 AGENT_KEY 存在
require_env CONVERSATION_ID "set CONVERSATION_ID before resuming chat."  # 确保 CONVERSATION_ID 存在

# 发送恢复对话请求，-N 参数禁用缓冲。
af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/resume" \
  -H "Content-Type: application/json" \
  -d '{"conversation_id":"'"$CONVERSATION_ID"'"}'  # 发送请求体
echo  # 输出空行
