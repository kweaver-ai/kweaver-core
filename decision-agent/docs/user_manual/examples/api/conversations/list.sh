#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询某个 Agent 应用下的对话列表。
# 请求：GET /api/agent-factory/v1/app/{agent_key}/conversation
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用
require_env AGENT_KEY "set AGENT_KEY before listing conversations."  # 确保 AGENT_KEY 存在

af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation?page=${PAGE:-1}&size=${LIMIT:-10}" | jq .  # 发送查询请求并美化输出
