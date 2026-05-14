#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：发起一次非流式对话，一次性返回完整 JSON。
# 请求：POST /api/agent-factory/v1/app/{agent_key}/chat/completion
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
ensure_agent_id  # 确保 AGENT_ID 存在
require_env AGENT_KEY "set AGENT_KEY from a published Agent."  # 确保 AGENT_KEY 存在

# 发送非流式对话请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"'"$AGENT_ID"'","agent_key":"'"$AGENT_KEY"'","agent_version":"'"${AGENT_VERSION:-v1}"'","query":"'"${CHAT_MESSAGE:-请用一句话介绍你自己}"'","stream":false}'  # 发送请求体，禁用流式
echo  # 输出空行
