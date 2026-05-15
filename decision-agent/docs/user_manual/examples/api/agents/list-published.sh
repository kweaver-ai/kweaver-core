#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询已发布到广场的 Agent 列表。
# 请求：POST /api/agent-factory/v3/published/agent
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

LIMIT="${LIMIT:-1}"  # 设置查询数量限制，默认为1

# 使用 jq 创建 JSON 对象。
jq -n \
  --argjson limit "$LIMIT" \
  '{
    offset: 0,
    limit: $limit,
    category_id: "",
    name: "",
    custom_space_id: "",
    is_to_square: 1
  }' \
  | af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/published/agent" \
      -H "Content-Type: text/plain;charset=UTF-8" \
      --data-binary @- \
  | jq .
