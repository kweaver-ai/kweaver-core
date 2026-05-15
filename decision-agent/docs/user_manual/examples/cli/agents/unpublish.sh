#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：取消发布当前 Agent。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在

kweaver agent unpublish "$AGENT_ID"  # 调用kweaver命令取消发布Agent
