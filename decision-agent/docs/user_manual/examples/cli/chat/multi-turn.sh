#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：使用 kweaver agent chat 演示同一个对话内的多轮问答。
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_kweaver  # 检查 kweaver 命令是否可用
ensure_agent_id  # 确保 AGENT_ID 存在

WORK_DIR="$EXAMPLES_STATE_DIR/cli-chat"  # 设置工作目录路径
mkdir -p "$WORK_DIR"  # 创建工作目录

FIRST_MESSAGE="${MULTI_TURN_FIRST_MESSAGE:-${CHAT_MESSAGE:-请用一句话介绍你自己}}"  # 设置第一轮问题
SECOND_MESSAGE="${MULTI_TURN_SECOND_MESSAGE:-请基于上一轮回答，再补充一个使用建议}"  # 设置第二轮问题

FIRST_STDOUT="$WORK_DIR/multi-turn-first.out"  # 第一轮标准输出文件
FIRST_STDERR="$WORK_DIR/multi-turn-first.err"  # 第一轮标准错误文件
SECOND_STDOUT="$WORK_DIR/multi-turn-second.out"  # 第二轮标准输出文件
SECOND_STDERR="$WORK_DIR/multi-turn-second.err"  # 第二轮标准错误文件

echo "第一轮问题：$FIRST_MESSAGE" >&2
# 第一轮非流式对话；--verbose 会在 stderr 中输出 conversation_id。
kweaver agent chat "$AGENT_ID" \
  --version "${AGENT_VERSION:-v0}" \
  --message "$FIRST_MESSAGE" \
  --no-stream \
  --verbose \
  > >(tee "$FIRST_STDOUT") \
  2> "$FIRST_STDERR"

CONVERSATION_ID="$(
  sed -n 's/^conversation_id:[[:space:]]*//p' "$FIRST_STDERR" \
    | tail -n 1
)"  # 从第一轮 verbose 输出中提取对话 ID

examples_require_env CONVERSATION_ID "first chat response did not include conversation_id."  # 确保对话 ID 存在
examples_state_set CONVERSATION_ID "$CONVERSATION_ID"  # 保存对话 ID，供 history/trace 等示例复用

echo "第二轮问题：$SECOND_MESSAGE" >&2
echo "继续对话：$CONVERSATION_ID" >&2
# 第二轮带上 conversation_id，继续同一个对话。
kweaver agent chat "$AGENT_ID" \
  --version "${AGENT_VERSION:-v0}" \
  --conversation-id "$CONVERSATION_ID" \
  --message "$SECOND_MESSAGE" \
  --no-stream \
  --verbose \
  > >(tee "$SECOND_STDOUT") \
  2> "$SECOND_STDERR"

echo "已将 CONVERSATION_ID=$CONVERSATION_ID 保存到 $EXAMPLES_STATE_FILE" >&2
