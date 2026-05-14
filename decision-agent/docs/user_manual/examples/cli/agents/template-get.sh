#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：查询单个 Agent 模板详情。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
skip_if_empty TEMPLATE_ID "set TEMPLATE_ID before running kweaver agent template-get."  # 检查TEMPLATE_ID是否为空

kweaver agent template-get "$TEMPLATE_ID" --pretty  # 调用kweaver命令查询指定模板的详情，以美化格式输出
