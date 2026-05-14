#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：给 Agent 绑定技能。默认跳过，避免依赖本地不存在的 SKILL_ID。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在
skip_if_empty SKILL_ID "set SKILL_ID before adding a skill."  # 检查SKILL_ID是否为空

kweaver agent skill add "$AGENT_ID" "$SKILL_ID"  # 调用kweaver命令为Agent添加技能
