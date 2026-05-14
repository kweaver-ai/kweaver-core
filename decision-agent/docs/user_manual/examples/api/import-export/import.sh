#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：导入上一步导出的 Agent 文件。
# 请求：POST /api/agent-factory/v3/agent-inout/import
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

EXPORT_FILE="${EXPORT_FILE:-$API_TMP_DIR/agent-export-result.json}"  # 设置导出文件路径
if [[ ! -f "$EXPORT_FILE" ]]; then  # 检查导出文件是否存在
  echo "Error: export file not found. Run make export first or set EXPORT_FILE." >&2  # 输出错误信息
  exit 1  # 退出
fi

# 发送导入请求。
af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/import" \
  -F "import_type=${IMPORT_TYPE:-upsert}" \
  -F "file=@$EXPORT_FILE;type=application/json" \
  -o "$API_TMP_DIR/agent-import-result.json"  # 保存响应到文件

jq . "$API_TMP_DIR/agent-import-result.json"  # 美化输出响应
