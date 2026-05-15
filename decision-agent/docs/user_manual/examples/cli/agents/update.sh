#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：更新 Agent 基础信息；如设置 KN_ID，则同时演示知识网络绑定。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在

UPDATED_NAME="${AGENT_NAME:-example_cli_agent_updated_$(date +%Y%m%d%H%M%S)}"  # 设置更新后的Agent名称，默认使用时间戳
UPDATED_PROFILE="${AGENT_PROFILE:-Updated by docs/user_manual/examples/cli}"  # 设置更新后的Agent描述，默认为固定值

if [[ -n "${KN_ID:-}" ]]; then  # 检查是否设置了知识网络ID
  # 调用 kweaver 更新 Agent 命令，包含知识网络。
  kweaver agent update "$AGENT_ID" \
    --name "$UPDATED_NAME" \
    --profile "$UPDATED_PROFILE" \
    --knowledge-network-id "$KN_ID" \
    --pretty  # 以美化格式输出
else  # 如果未设置知识网络ID
  # 调用 kweaver 更新 Agent 命令，不包含知识网络。
  kweaver agent update "$AGENT_ID" \
    --name "$UPDATED_NAME" \
    --profile "$UPDATED_PROFILE" \
    --pretty  # 以美化格式输出
fi
