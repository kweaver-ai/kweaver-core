#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：串起 kweaver agent create/get/update/publish/unpublish/delete 完整流程。
# 注意：该脚本会创建并删除临时 Agent，不属于 quick-check。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库

cleanup() {  # 定义清理函数
  if [[ -n "$AGENT_ID" ]]; then  # 检查AGENT_ID是否存在
    "$(dirname "${BASH_SOURCE[0]}")/unpublish.sh" >/dev/null 2>&1 || true  # 调用取消发布脚本，忽略错误
    "$(dirname "${BASH_SOURCE[0]}")/delete.sh" >/dev/null 2>&1 || true  # 调用删除脚本，忽略错误
  fi
}
trap cleanup EXIT  # 设置退出陷阱，确保脚本退出时执行清理函数

"$(dirname "${BASH_SOURCE[0]}")/create.sh"  # 执行创建脚本
examples_load_kv_file "$EXAMPLES_STATE_FILE"  # 加载状态文件中的环境变量
examples_require_env AGENT_ID "create target did not save AGENT_ID."  # 确保AGENT_ID存在

"$(dirname "${BASH_SOURCE[0]}")/get.sh"  # 执行获取脚本
"$(dirname "${BASH_SOURCE[0]}")/update.sh"  # 执行更新脚本
"$(dirname "${BASH_SOURCE[0]}")/publish.sh"  # 执行发布脚本
"$(dirname "${BASH_SOURCE[0]}")/unpublish.sh"  # 执行取消发布脚本
"$(dirname "${BASH_SOURCE[0]}")/delete.sh"  # 执行删除脚本
AGENT_ID=""  # 清空AGENT_ID变量
echo "CLI flow finished."  # 输出流程完成信息
