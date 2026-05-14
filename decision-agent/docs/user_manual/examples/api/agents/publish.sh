#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：发布 Agent 到广场。
# 请求：POST /api/agent-factory/v3/agent/{agent_id}/publish
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用
ensure_agent_id  # 确保 AGENT_ID 存在

# 使用 jq 创建发布请求 JSON。
jq -n '{
  category_ids: [],
  description: "Published by docs example",
  publish_to_where: ["square"],
  publish_to_bes: ["skill_agent"],
  pms_control: null
}' > "$API_TMP_DIR/agent-publish.json"  # 生成发布请求文件

# 发送发布请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish" \
  -H "Content-Type: application/json" \
  -d @"$API_TMP_DIR/agent-publish.json" \
  | tee "$API_TMP_DIR/publish-response.json" \
  | jq .

AGENT_VERSION="$(jq -r '.version // empty' "$API_TMP_DIR/publish-response.json")"  # 从响应中提取 Agent 版本
[[ -n "$AGENT_VERSION" ]] && examples_state_set AGENT_VERSION "$AGENT_VERSION"  # 如果版本存在则保存到状态文件
