#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：删除指定对话。
# 请求：DELETE /api/agent-factory/v1/app/{agent_key}/conversation/{conversation_id}
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_env AGENT_KEY "set AGENT_KEY before deleting conversation."  # 确保 AGENT_KEY 存在
require_env CONVERSATION_ID "set CONVERSATION_ID before deleting conversation."  # 确保 CONVERSATION_ID 存在

af_curl -X DELETE "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID"  # 发送删除对话请求
echo  # 输出空行
