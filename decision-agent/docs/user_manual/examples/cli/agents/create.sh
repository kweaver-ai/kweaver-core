#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：创建一个最小可用 Agent，并保存 AGENT_ID 到 examples/.tmp/state.env。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查kweaver命令是否可用
require_jq  # 检查jq命令是否可用

WORK_DIR="$EXAMPLES_STATE_DIR/cli-agents"  # 设置工作目录路径
mkdir -p "$WORK_DIR"  # 创建工作目录，如果已存在则不报错

CONFIG_EXAMPLE_FILE="$CLI_EXAMPLE_DIR/agents/agent-config.json.example"  # 配置示例文件路径
AGENT_CONFIG_FILE="${AGENT_CONFIG_FILE:-$CLI_EXAMPLE_DIR/agents/agent-config.json}"  # 设置Agent配置文件路径，可通过环境变量覆盖

if [[ ! -f "$AGENT_CONFIG_FILE" ]]; then  # 检查配置文件是否存在
  cp "$CONFIG_EXAMPLE_FILE" "$AGENT_CONFIG_FILE"  # 从示例文件复制配置文件
  echo "已从 agent-config.json.example 初始化 $AGENT_CONFIG_FILE；如需自定义配置，请修改后重新运行。" >&2  # 输出提示信息到标准错误
fi

jq . "$AGENT_CONFIG_FILE" >/dev/null  # 验证配置文件是否为有效的JSON格式

AGENT_NAME="${AGENT_NAME:-example_cli_agent_$(date +%Y%m%d%H%M%S)}"  # 设置Agent名称，默认使用时间戳生成唯一名称

# 调用 kweaver 创建 Agent 命令。
kweaver agent create \
  --name "$AGENT_NAME" \
  --profile "Created by docs/user_manual/examples/cli" \
  --product-key dip \
  --config "$AGENT_CONFIG_FILE" \
  --pretty \
  | tee "$WORK_DIR/create-response.json"

AGENT_ID="$(jq -r '.id // empty' "$WORK_DIR/create-response.json")"  # 从响应中提取Agent ID
AGENT_VERSION="$(jq -r '.version // "v0"' "$WORK_DIR/create-response.json")"  # 从响应中提取Agent版本，默认v0
examples_require_env AGENT_ID "创建响应中没有返回 agent id。"  # 确保AGENT_ID存在
examples_state_set AGENT_ID "$AGENT_ID"  # 保存AGENT_ID到状态文件
examples_state_set AGENT_VERSION "$AGENT_VERSION"  # 保存AGENT_VERSION到状态文件

echo "已将 AGENT_ID=$AGENT_ID 保存到 $EXAMPLES_STATE_FILE" >&2  # 输出保存成功信息到标准错误
