#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：导出后再导入，并在导入成功时清理临时导入出的 Agent。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

"$(dirname "${BASH_SOURCE[0]}")/export.sh"  # 执行导出脚本
"$(dirname "${BASH_SOURCE[0]}")/import.sh"  # 执行导入脚本

IMPORTED_AGENT_ID="$(jq -r '.id // .data.id // .agent_id // .data.agent_id // empty' "$API_TMP_DIR/agent-import-result.json")"  # 从响应中提取导入的 Agent ID
if [[ "${CLEANUP_IMPORTED_AGENT:-1}" == "1" && -n "$IMPORTED_AGENT_ID" && "$IMPORTED_AGENT_ID" != "${AGENT_ID:-}" ]]; then  # 如果需要清理且导入的 Agent ID 与原 ID 不同
  echo "Cleaning imported agent: $IMPORTED_AGENT_ID"  # 输出清理信息
  AGENT_ID="$IMPORTED_AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/../agents/delete.sh"  # 删除导入的 Agent
fi
