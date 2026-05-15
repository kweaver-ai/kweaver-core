#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：为 Agent 配置知识网络。CLI 只传 knowledge_network_id。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
ensure_agent_id  # 确保AGENT_ID存在
skip_if_empty KN_ID "set KN_ID before updating knowledge network."  # 检查KN_ID是否为空

kweaver agent update "$AGENT_ID" --knowledge-network-id "$KN_ID" --pretty  # 调用kweaver命令更新Agent的知识网络ID，以美化格式输出
