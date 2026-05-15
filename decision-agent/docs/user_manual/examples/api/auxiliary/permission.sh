#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询当前用户对 Agent 管理能力的权限状态。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-permission/management/user-status" | jq .  # 发送查询用户权限状态请求并美化输出
