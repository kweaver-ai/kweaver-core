#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：删除当前 Agent。删除成功后清空 state 中的 Agent 信息。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
examples_require_env AGENT_ID "set AGENT_ID, add it to .env, or run make create before deleting."  # 确保AGENT_ID存在

kweaver agent delete "$AGENT_ID" -y  # 调用kweaver删除Agent命令，-y参数表示自动确认
STATE_AGENT_ID="$(grep -E '^AGENT_ID=' "$EXAMPLES_STATE_FILE" 2>/dev/null | tail -n 1 | cut -d= -f2- || true)"  # 从状态文件中获取保存的AGENT_ID
if [[ "$STATE_AGENT_ID" == "$AGENT_ID" ]]; then  # 检查状态文件中的AGENT_ID是否与要删除的ID一致
  examples_state_clear AGENT_ID AGENT_KEY AGENT_VERSION  # 清空状态文件中的Agent相关信息
fi
