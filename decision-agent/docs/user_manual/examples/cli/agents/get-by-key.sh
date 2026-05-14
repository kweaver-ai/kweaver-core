#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：按 agent_key 查询已发布 Agent 信息。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
skip_if_empty AGENT_KEY "set AGENT_KEY before running kweaver agent get-by-key."  # 检查AGENT_KEY是否为空

kweaver agent get-by-key "$AGENT_KEY"  # 调用kweaver命令根据agent_key查询Agent信息
