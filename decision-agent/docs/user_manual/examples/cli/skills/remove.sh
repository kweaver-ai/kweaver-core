#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：移除 Agent 技能绑定。默认跳过，避免误删未指定资源。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在
skip_if_empty SKILL_ID "set SKILL_ID before removing a skill."  # 检查SKILL_ID是否为空

kweaver agent skill remove "$AGENT_ID" "$SKILL_ID"  # 调用kweaver命令移除Agent的技能绑定
