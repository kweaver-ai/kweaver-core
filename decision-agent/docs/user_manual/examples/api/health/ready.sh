#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：检查 Decision Agent 后端是否已启动并可接收请求。
# 请求：GET /health/ready
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库

af_curl "$AF_BASE_URL/health/ready"  # 发送健康检查请求
echo  # 输出空行
