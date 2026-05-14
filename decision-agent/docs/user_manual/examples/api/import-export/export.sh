#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：导出指定 Agent。
# 请求：POST /api/agent-factory/v3/agent-inout/export
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用
ensure_agent_id  # 确保 AGENT_ID 存在

jq -n --arg agent_id "$AGENT_ID" '{agent_ids: [$agent_id]}' > "$API_TMP_DIR/agent-export.json"  # 生成导出请求 JSON

# 发送导出请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/export" \
  -H "Content-Type: application/json" \
  -d @"$API_TMP_DIR/agent-export.json" \
  -o "$API_TMP_DIR/agent-export-result.json"  # 保存响应到文件

echo "Exported to api/.tmp/agent-export-result.json"  # 输出导出成功信息
