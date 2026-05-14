#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：使用 kweaver agent chat 发起非流式对话。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在

# 调用 kweaver 与 Agent 对话命令。
kweaver agent chat "$AGENT_ID" \
  --version "${AGENT_VERSION:-v0}" \
  --message "${CHAT_MESSAGE:-请用一句话介绍你自己}" \
  --no-stream \
  --verbose  # 显示详细信息
