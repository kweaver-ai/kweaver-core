#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

API_EXAMPLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"  # 获取 API 示例目录的绝对路径
API_TMP_DIR="$API_EXAMPLE_DIR/.tmp"  # 设置 API 临时目录路径
source "$API_EXAMPLE_DIR/../_env.sh"  # 加载环境变量配置文件
examples_load_env  # 加载示例所需的环境变量

# 接口示例默认连接本地 Decision Agent；如需访问其他环境，可在运行前覆盖这些环境变量。
AF_BASE_URL="${AF_BASE_URL:-$KWEAVER_BASE_URL}"  # 设置 Agent Factory 基础 URL
AF_BD="${AF_BD:-$KWEAVER_BUSINESS_DOMAIN}"  # 设置业务域
AF_TOKEN="${AF_TOKEN:-${KWEAVER_TOKEN:-__NO_AUTH__}}"  # 设置认证令牌，默认为无认证

mkdir -p "$API_TMP_DIR"  # 创建 API 临时目录

af_curl() {  # 定义函数：封装 curl 调用 Agent Factory API
  # 本地免鉴权模式下不发送 Authorization / Token，只保留业务域和语言请求头。
  if [[ "${KWEAVER_NO_AUTH:-}" == "1" || "$AF_TOKEN" == "__NO_AUTH__" ]]; then  # 如果是无认证模式
    curl -fsS -H "X-Business-Domain: $AF_BD" -H "X-Language: zh-cn" "$@"  # 发送不带认证的请求
  else  # 如果需要认证
    # 发送带认证的请求。
    curl -fsS \
      -H "Authorization: Bearer $AF_TOKEN" \
      -H "Token: $AF_TOKEN" \
      -H "X-Business-Domain: $AF_BD" \
      -H "X-Language: zh-cn" \
      "$@"
  fi
}

require_jq() {  # 定义函数：检查 jq 命令是否可用
  if ! command -v jq >/dev/null 2>&1; then  # 检查 jq 命令是否存在
    echo "jq is required for this example." >&2  # 输出错误信息到标准错误
    exit 1  # 退出脚本，返回错误码1
  fi
}

require_env() {  # 定义函数：要求环境变量存在
  local name="$1"  # 获取第一个参数：变量名
  local message="$2"  # 获取第二个参数：错误信息
  examples_require_env "$name" "$message"  # 调用环境变量检查函数
}

ensure_agent_id() {  # 定义函数：确保 AGENT_ID 存在
  examples_ensure_agent_id "$API_EXAMPLE_DIR/agents/create.sh"  # 调用确保 AGENT_ID 的函数
}
