#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 命令行示例默认从本目录 node_modules/.bin 取 kweaver，确保本地安装包和文档一致。
CLI_EXAMPLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"  # 获取当前脚本所在目录的绝对路径
source "$CLI_EXAMPLE_DIR/../_env.sh"  # 加载环境变量配置文件
examples_load_env  # 加载示例所需的环境变量

export PATH="$CLI_EXAMPLE_DIR/node_modules/.bin:$PATH"  # 将本地node_modules/.bin目录添加到PATH环境变量

require_kweaver() {  # 定义函数：检查kweaver命令是否可用
  if ! command -v kweaver >/dev/null 2>&1; then  # 检查kweaver命令是否存在
    echo "kweaver command is required. Run make install in docs/user_manual/examples/cli first." >&2  # 输出错误信息到标准错误
    exit 1  # 退出脚本，返回错误码1
  fi
}

require_jq() {  # 定义函数：检查jq命令是否可用
  if ! command -v jq >/dev/null 2>&1; then  # 检查jq命令是否存在
    echo "jq is required for this example." >&2  # 输出错误信息到标准错误
    exit 1  # 退出脚本，返回错误码1
  fi
}

skip_if_empty() {  # 定义函数：如果环境变量为空则跳过执行
  local var_name="$1"  # 获取第一个参数：变量名
  local message="$2"  # 获取第二个参数：提示信息
  examples_require_env "$var_name" "$message"  # 调用环境变量检查函数
}

ensure_agent_id() {  # 定义函数：确保AGENT_ID存在
  examples_ensure_agent_id "$CLI_EXAMPLE_DIR/agents/create.sh"  # 调用确保AGENT_ID的函数
}
