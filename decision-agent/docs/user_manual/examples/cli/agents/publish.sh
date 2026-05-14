#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：发布当前 Agent。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
require_jq  # 检查jq命令是否可用
ensure_agent_id  # 确保AGENT_ID存在

WORK_DIR="$EXAMPLES_STATE_DIR/cli-agents"  # 设置工作目录路径
mkdir -p "$WORK_DIR"  # 创建工作目录，如果已存在则不报错

kweaver agent publish "$AGENT_ID" | tee "$WORK_DIR/publish-response.json"  # 调用kweaver发布Agent命令，并将响应同时输出到控制台和文件

AGENT_VERSION="$(jq -r '.version // empty' "$WORK_DIR/publish-response.json")"  # 从响应中提取Agent版本号
[[ -n "$AGENT_VERSION" ]] && examples_state_set AGENT_VERSION "$AGENT_VERSION"  # 如果版本号存在则保存到状态文件
