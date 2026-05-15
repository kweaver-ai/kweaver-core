#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：创建一个最小可用 Agent。
# 前置：设置 KWEAVER_LLM_ID / KWEAVER_LLM_NAME。
# 请求：POST /api/agent-factory/v3/agent
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

: "${KWEAVER_LLM_ID:?Please set KWEAVER_LLM_ID.}"  # 检查 KWEAVER_LLM_ID 是否设置
: "${KWEAVER_LLM_NAME:?Please set KWEAVER_LLM_NAME.}"  # 检查 KWEAVER_LLM_NAME 是否设置

CONFIG_FILE="$API_TMP_DIR/agent-config-minimal.json"  # 设置配置文件路径
CREATE_FILE="$API_TMP_DIR/create-agent.json"  # 设置创建请求文件路径
RESPONSE_FILE="$API_TMP_DIR/create-response.json"  # 设置响应文件路径
AGENT_NAME="${AGENT_NAME:-example_api_agent_$(date +%Y%m%d%H%M%S)}"  # 设置 Agent 名称，默认使用时间戳

"$(dirname "${BASH_SOURCE[0]}")/../agent-config/minimal.sh" >/dev/null  # 生成最小配置

# 使用 jq 创建 JSON 对象。
jq -n \
  --arg name "$AGENT_NAME" \
  --argjson config "$(cat "$CONFIG_FILE")" \
  '{
    name: $name,
    profile: "Created by docs/user_manual/examples/api",
    avatar_type: 1,
    avatar: "icon-dip-agent-default",
    product_key: "DIP",
    config: $config
  }' > "$CREATE_FILE"  # 生成创建请求 JSON

# 发送创建 Agent 请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent" \
  -H "Content-Type: application/json" \
  -d @"$CREATE_FILE" \
  | tee "$RESPONSE_FILE" \
  | jq .

AGENT_ID="$(jq -r '.id // empty' "$RESPONSE_FILE")"  # 从响应中提取 Agent ID
AGENT_VERSION="$(jq -r '.version // "v0"' "$RESPONSE_FILE")"  # 从响应中提取 Agent 版本
examples_require_env AGENT_ID "create response did not include agent id."  # 确保 AGENT_ID 存在
examples_state_set AGENT_ID "$AGENT_ID"  # 保存 AGENT_ID 到状态文件
examples_state_set AGENT_VERSION "$AGENT_VERSION"  # 保存 AGENT_VERSION 到状态文件

echo "Saved response to api/.tmp/create-response.json" >&2  # 输出响应保存位置
echo "Saved AGENT_ID=$AGENT_ID to $EXAMPLES_STATE_FILE" >&2  # 输出 AGENT_ID 保存位置
