#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询广场 Agent 列表。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用

kweaver agent list --limit "${LIMIT:-1}" --pretty  # 调用kweaver命令查询广场Agent列表，限制数量默认为1，以美化格式输出
