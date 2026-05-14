#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询 Agent 的对话列表。sessions 是 CLI 历史命令名。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在

kweaver agent sessions "$AGENT_ID" --limit "${LIMIT:-10}" --pretty  # 调用kweaver命令查询对话列表，限制数量默认为10，以美化格式输出
