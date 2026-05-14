#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：发起增量流式对话。每个 SSE data 是增量 patch：
# { "seq_id": 1, "key": ["message", "content"], "content": "...", "action": "append" }
# 客户端需要按 seq_id 顺序消费，并根据 key/action 合并到本地响应对象。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
ensure_agent_id  # 确保 AGENT_ID 存在
require_env AGENT_KEY "set AGENT_KEY from a published Agent."  # 确保 AGENT_KEY 存在

# 发送增量流式对话请求，-N 参数禁用缓冲。
af_curl -N -X POST "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"'"$AGENT_ID"'","agent_key":"'"$AGENT_KEY"'","agent_version":"'"${AGENT_VERSION:-v1}"'","query":"'"${CHAT_MESSAGE:-请用一句话介绍你自己}"'","stream":true,"inc_stream":true}'  # 发送请求体，启用增量流
echo  # 输出空行
