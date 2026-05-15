#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：串起创建 -> 查看 -> 发布 -> 取消发布 -> 删除的完整低风险清理流程。
# 注意：该脚本会创建并删除临时 Agent，不属于 quick-check。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

: "${KWEAVER_LLM_ID:?Please set KWEAVER_LLM_ID before running make flow.}"  # 检查 KWEAVER_LLM_ID 是否设置
: "${KWEAVER_LLM_NAME:?Please set KWEAVER_LLM_NAME before running make flow.}"  # 检查 KWEAVER_LLM_NAME 是否设置

AGENT_ID=""  # 初始化 AGENT_ID 为空
cleanup() {  # 定义清理函数
  if [[ -n "$AGENT_ID" ]]; then  # 如果 AGENT_ID 存在
    AGENT_ID="$AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/unpublish.sh" >/dev/null 2>&1 || true  # 取消发布，忽略错误
    AGENT_ID="$AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/delete.sh" >/dev/null 2>&1 || true  # 删除 Agent，忽略错误
  fi
}
trap cleanup EXIT  # 设置退出陷阱，确保脚本退出时执行清理函数

"$(dirname "${BASH_SOURCE[0]}")/create.sh"  # 执行创建脚本
AGENT_ID="$(jq -r '.id // empty' "$API_TMP_DIR/create-response.json")"  # 从响应中提取 Agent ID
test -n "$AGENT_ID"  # 确保 AGENT_ID 不为空

AGENT_ID="$AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/get.sh"  # 执行获取脚本
AGENT_ID="$AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/publish.sh"  # 执行发布脚本
AGENT_ID="$AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/unpublish.sh"  # 执行取消发布脚本
AGENT_ID="$AGENT_ID" "$(dirname "${BASH_SOURCE[0]}")/delete.sh"  # 执行删除脚本
AGENT_ID=""  # 清空 AGENT_ID

echo "API flow finished."  # 输出流程完成信息
