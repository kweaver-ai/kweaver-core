#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：按 agent_id 获取 Agent 详情。
# 请求：GET /api/agent-factory/v3/agent/{agent_id}
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用
ensure_agent_id  # 确保 AGENT_ID 存在

RESPONSE_FILE="$API_TMP_DIR/get-response.json"  # 设置响应文件路径
af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" | tee "$RESPONSE_FILE" | jq .  # 发送获取请求并美化输出

AGENT_KEY="$(jq -r '.key // empty' "$RESPONSE_FILE")"  # 从响应中提取 Agent Key
AGENT_VERSION="$(jq -r '.version // .published_version // "v0"' "$RESPONSE_FILE")"  # 从响应中提取 Agent 版本
[[ -n "$AGENT_KEY" ]] && examples_state_set AGENT_KEY "$AGENT_KEY"  # 如果 Agent Key 存在则保存到状态文件
[[ -n "$AGENT_VERSION" ]] && examples_state_set AGENT_VERSION "$AGENT_VERSION"  # 如果 Agent 版本存在则保存到状态文件
