#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：删除 Agent。
# 请求：DELETE /api/agent-factory/v3/agent/{agent_id}
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_env AGENT_ID "set AGENT_ID before deleting."  # 确保 AGENT_ID 存在

af_curl -X DELETE "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"  # 发送删除 Agent 请求
STATE_AGENT_ID="$(grep -E '^AGENT_ID=' "$EXAMPLES_STATE_FILE" 2>/dev/null | tail -n 1 | cut -d= -f2- || true)"  # 从状态文件中获取保存的 AGENT_ID
if [[ "$STATE_AGENT_ID" == "$AGENT_ID" ]]; then  # 检查状态文件中的 AGENT_ID 是否与要删除的 ID 一致
  examples_state_clear AGENT_ID AGENT_KEY AGENT_VERSION  # 清空状态文件中的 Agent 相关信息
fi
echo  # 输出空行
