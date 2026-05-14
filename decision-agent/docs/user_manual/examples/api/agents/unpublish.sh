#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：取消发布 Agent。
# 请求：PUT /api/agent-factory/v3/agent/{agent_id}/unpublish
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
ensure_agent_id  # 确保 AGENT_ID 存在

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish"  # 发送取消发布请求
echo  # 输出空行
